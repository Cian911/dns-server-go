package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	addr := "127.0.0.1:2053"
	resolver := ""

	if len(os.Args) > 1 && os.Args[1] == "--resolver" && os.Args[2] != "" {
		resolver = os.Args[2]
	}

	fmt.Printf("Resolver is %s\n", resolver)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	for {
		buf := make([]byte, 2056)

		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		fmt.Printf("Received %d bytes from %s\n", size, source)

		if len(resolver) != 0 {
			mainQuery := NewQuery(buf[:size])
			questionCount := len(mainQuery.Question)
			answers := make([]Answer, 0)
			fmt.Printf("QCOUNT: %d\n", questionCount)

			ip, port := parseAddr(resolver)
			forwardAddr := net.UDPAddr{
				IP:   net.ParseIP(ip),
				Port: port,
			}

			for i, question := range mainQuery.Question {
				singleQuery := &Message{
					Header: Header{
						ID:      mainQuery.Header.ID,
						QR:      0,
						OPCODE:  0,
						AA:      0,
						TC:      0,
						RD:      1,
						RA:      0,
						Z:       0,
						RCODE:   0,
						QDCOUNT: 1,
						ANCOUNT: 0,
						NSCOUNT: 0,
						ARCOUNT: 0,
					},
					Question:       []Question{question},
					ResourceRecord: nil,
					Answer:         nil,
				}

				queryBytes := singleQuery.Bytes()
				queryBytes = stripTrailingZeroValues(queryBytes)

				forwardConn, err := net.DialUDP("udp", nil, &forwardAddr)
				if err != nil {
					log.Fatal(err)
				}
				defer forwardConn.Close()

				fmt.Printf("Forwarding question %d: %s\n", i, string(question.Name))
				_, err = forwardConn.Write(queryBytes)
				if err != nil {
					log.Fatalf("Error forwarding with single question: %v", err)
				}

				forwardConn.SetReadDeadline(time.Now().Add(10 * time.Second))
				resolverBuffer := make([]byte, 2056)
				resolverSize, _, err := forwardConn.ReadFromUDP(resolverBuffer)
				if err != nil {
					log.Fatalf("Error reading from resolver: %v", err)
				}

				fmt.Printf("Received %d bytes from resolver\n", resolverSize)

				responseQuery := NewQuery(resolverBuffer[:resolverSize])
				answers = append(answers, responseQuery.Answer...)
				i += 1
			}

			response := &Message{
				Header: Header{
					ID:      mainQuery.Header.ID,
					QR:      1,
					OPCODE:  mainQuery.Header.OPCODE,
					AA:      0,
					TC:      0,
					RD:      mainQuery.Header.RD,
					RA:      0,
					Z:       0,
					RCODE:   4,
					QDCOUNT: uint16(questionCount),
					ANCOUNT: uint16(len(answers)),
					NSCOUNT: 0,
					ARCOUNT: 0,
				},
				Question: mainQuery.Question,
				Answer:   answers,
			}
			fmt.Println("RESOLVER RESPONSE...")

			responseBytes := response.Bytes()
			responseBytes = stripTrailingZeroValues(responseBytes)

			_, err = udpConn.WriteToUDP(responseBytes, source)
			if err != nil {
				fmt.Println("Failed to send response:", err)
			}
		} else {
			fmt.Printf("No resolver specified, echoing back\n")
			fmt.Println("NO RESOLVER SPECIFIED...")
			m := NewQuery(buf[:size])

			_, err = udpConn.WriteToUDP(m.Bytes(), source)
			if err != nil {
				fmt.Println("Failed to send response:", err)
			}
		}
	}
}

func stripTrailingZeroValues(input []byte) []byte {
	// Find the last non-zero element
	lastNonZeroIndex := len(input) - 1
	for lastNonZeroIndex >= 0 && input[lastNonZeroIndex] == 0 {
		lastNonZeroIndex--
	}
	// Return a slice of the array up to the last non-zero element
	return input[:lastNonZeroIndex+1]
}
