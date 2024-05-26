// package main
//
// import (
// 	"encoding/binary"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"time"
// )
//
// func main() {
// 	// You can use print statements as follows for debugging, they'll be visible when running tests.
// 	fmt.Println("Logs from your program will appear here!")
// 	addr := "127.0.0.1:2053"
// 	resolver := ""
//
// 	if len(os.Args) > 1 && os.Args[1] == "--resolver" && os.Args[2] != "" {
// 		resolver = os.Args[2]
// 	}
//
// 	fmt.Printf("Resolver is %s\n", resolver)
//
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	if err != nil {
// 		fmt.Println("Failed to resolve UDP address:", err)
// 		return
// 	}
//
// 	udpConn, err := net.ListenUDP("udp", udpAddr)
// 	if err != nil {
// 		fmt.Println("Failed to bind to address:", err)
// 		return
// 	}
// 	defer udpConn.Close()
//
// 	buf := make([]byte, 1024)
//
// 	for {
//
// 		size, source, err := udpConn.ReadFromUDP(buf)
// 		if err != nil {
// 			fmt.Println("Error receiving data:", err)
// 			break
// 		}
//
// 		receivedData := string(buf[:size])
//
// 		if len(resolver) != 0 {
//
// 			// Check if query contains multiple questions
// 			questionCount := int(binary.BigEndian.Uint16(buf[4:6]))
// 			questions := make([]Question, 0, questionCount)
// 			answers := make([]Answer, 0, questionCount)
//
// 			if questionCount > 1 {
// 				// Split into seperate queries
// 				mainQuery := NewQuery(buf[:size])
// 				fmt.Println("Appending questions...")
//
// 				for _, v := range mainQuery.Question {
// 					questions = append(questions, v)
// 				}
// 			}
//
// 			resolverBuffer := make([]byte, 1024)
// 			ip, port := parseAddr(resolver)
// 			forwardAddr := net.UDPAddr{
// 				IP:   net.ParseIP(ip),
// 				Port: port,
// 			}
//
// 			// Connect to resolver ip
// 			forwardConn, err := net.DialUDP("udp", nil, &forwardAddr)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			// Defer closing the connection
// 			defer forwardConn.Close()
//
// 			// Write to the connection
// 			if questionCount > 1 {
// 				for _, v := range questions {
// 					q := NewQuery(buf[:size])
// 					q.Question = []Question{v}
// 					_, err = forwardConn.Write(q.Bytes())
// 					if err != nil {
// 						log.Fatalf("Error forwarding with multi-question: %v", err)
// 					}
// 					// Set read deadline so we dont wait infinately
// 					forwardConn.SetReadDeadline(time.Now().Add(10 * time.Second))
// 					// Read from resolver connection
// 					resolverSize, resolverSource, err := forwardConn.ReadFromUDP(resolverBuffer)
// 					if err != nil {
// 						log.Fatal(err)
// 					}
// 					resolverData := string(resolverBuffer[:resolverSize])
// 					fmt.Printf("Received %d bytes from resolver %s: %s\n", resolverSize, resolverSource, resolverData)
//
// 					m := NewQuery(resolverBuffer[:resolverSize])
// 					fmt.Printf("M: %v", m.Question)
// 					answers = append(answers, m.Answer[0])
// 				}
// 			} else {
// 				_, err = forwardConn.Write(buf[:size])
// 			}
// 			if err != nil {
// 				log.Fatal(err)
// 			}
//
// 			// Set read deadline so we dont wait infinately
// 			forwardConn.SetReadDeadline(time.Now().Add(3 * time.Second))
// 			// Read from resolver connection
// 			resolverSize, resolverSource, err := forwardConn.ReadFromUDP(resolverBuffer)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
//
// 			resolverData := string(resolverBuffer[:resolverSize])
//
// 			fmt.Printf("Received %d bytes from resolver %s: %s\n", resolverSize, resolverSource, resolverData)
//
// 			m := NewQuery(resolverBuffer[:resolverSize])
// 			if len(answers) > 1 {
// 				m.Answer = answers
// 				answers = make([]Answer, 0)
// 			}
//
// 			// Write back to original requester
// 			_, err = udpConn.WriteToUDP(m.Bytes(), source)
// 			if err != nil {
// 				fmt.Println("Failed to send response:", err)
// 			}
// 		} else {
// 			fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)
//
// 			m := NewQuery(buf[:size])
//
// 			// Create an empty response
// 			_, err = udpConn.WriteToUDP(m.Bytes(), source)
// 			if err != nil {
// 				fmt.Println("Failed to send response:", err)
// 			}
// 		}
// 	}
// }

// package main
//
// import (
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"time"
// )
//
// func main() {
// 	fmt.Println("Logs from your program will appear here!")
// 	addr := "127.0.0.1:2053"
// 	resolver := ""
//
// 	if len(os.Args) > 1 && os.Args[1] == "--resolver" && os.Args[2] != "" {
// 		resolver = os.Args[2]
// 	}
//
// 	fmt.Printf("Resolver is %s\n", resolver)
//
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	if err != nil {
// 		fmt.Println("Failed to resolve UDP address:", err)
// 		return
// 	}
//
// 	udpConn, err := net.ListenUDP("udp", udpAddr)
// 	if err != nil {
// 		fmt.Println("Failed to bind to address:", err)
// 		return
// 	}
// 	defer udpConn.Close()
//
// 	buf := make([]byte, 1024)
//
// 	for {
// 		size, source, err := udpConn.ReadFromUDP(buf)
// 		if err != nil {
// 			fmt.Println("Error receiving data:", err)
// 			break
// 		}
//
// 		fmt.Printf("Received %d bytes from %s\n", size, source)
//
// 		if len(resolver) != 0 {
// 			mainQuery := NewQuery(buf[:size])
// 			questionCount := len(mainQuery.Question)
// 			answers := make([]Answer, 0)
//
// 			ip, port := parseAddr(resolver)
// 			forwardAddr := net.UDPAddr{
// 				IP:   net.ParseIP(ip),
// 				Port: port,
// 			}
//
// 			for i, question := range mainQuery.Question {
// 				singleQuery := &Message{
// 					Header: Header{
// 						ID:      mainQuery.Header.ID,
// 						QR:      0,
// 						OPCODE:  mainQuery.Header.OPCODE,
// 						AA:      0,
// 						TC:      0,
// 						RD:      mainQuery.Header.RD,
// 						RA:      0,
// 						Z:       0,
// 						RCODE:   0,
// 						QDCOUNT: 1,
// 						ANCOUNT: 0,
// 						NSCOUNT: 0,
// 						ARCOUNT: 0,
// 					},
// 					Question: []Question{question},
// 				}
//
// 				forwardConn, err := net.DialUDP("udp", nil, &forwardAddr)
// 				if err != nil {
// 					log.Fatal(err)
// 				}
// 				defer forwardConn.Close()
//
// 				fmt.Printf("Forwarding question %d: %s\n", i+1, string(question.Name))
// 				_, err = forwardConn.Write(singleQuery.Bytes())
// 				if err != nil {
// 					log.Fatalf("Error forwarding with single question: %v", err)
// 				}
//
// 				forwardConn.SetReadDeadline(time.Now().Add(10 * time.Second))
// 				resolverBuffer := make([]byte, 2056)
// 				resolverSize, _, err := forwardConn.ReadFromUDP(resolverBuffer)
// 				if err != nil {
// 					log.Fatalf("Error reading from resolver: %v", err)
// 				}
//
// 				fmt.Printf("Received %d bytes from resolver\n", resolverSize)
//
// 				responseQuery := NewQuery(resolverBuffer[:resolverSize])
// 				fmt.Println("ANS: ", string(responseQuery.Answer[0].Name))
// 				answers = append(answers, responseQuery.Answer...)
// 			}
//
// 			mainQuery.Header.ANCOUNT = uint16(questionCount)
// 			mainQuery.Header.QDCOUNT = uint16(questionCount)
// 			fmt.Println("HEADER ID: ", uint16(mainQuery.Header.ID))
// 			response := &Message{
// 				Header: Header{
// 					ID:      mainQuery.Header.ID,
// 					QR:      1,
// 					OPCODE:  0,
// 					AA:      0,
// 					TC:      0,
// 					RD:      1,
// 					RA:      0,
// 					Z:       0,
// 					RCODE:   4,
// 					QDCOUNT: uint16(questionCount),
// 					ANCOUNT: uint16(len(answers)),
// 					NSCOUNT: 0,
// 					ARCOUNT: 0,
// 				},
// 				Question: mainQuery.Question,
// 				Answer:   answers,
// 			}
//
// 			responseBytes := response.Bytes()
// 			fmt.Println(hex.Dump(response.Header.Bytes()))
// 			fmt.Println(hex.Dump(mainQuery.Header.Bytes()))
//
// 			_, err = udpConn.WriteToUDP(responseBytes, source)
// 			if err != nil {
// 				fmt.Println("Failed to send response:", err)
// 			}
// 		} else {
// 			fmt.Printf("No resolver specified, echoing back\n")
// 			m := NewQuery(buf[:size])
//
// 			_, err = udpConn.WriteToUDP(m.Bytes(), source)
// 			if err != nil {
// 				fmt.Println("Failed to send response:", err)
// 			}
// 		}
// 	}
// }

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
				fmt.Printf("Constructed query (%d bytes): %v\n", len(queryBytes), queryBytes)
				queryBytes = stripTrailingZeroValues(queryBytes)

				forwardConn, err := net.DialUDP("udp", nil, &forwardAddr)
				if err != nil {
					log.Fatal(err)
				}
				defer forwardConn.Close()

				fmt.Printf("Forwarding question %d: %s\n", i+1, string(question.Name))
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
			}

			response := &Message{
				Header: Header{
					ID:      mainQuery.Header.ID,
					QR:      1,
					OPCODE:  0,
					AA:      0,
					TC:      0,
					RD:      1,
					RA:      0,
					Z:       0,
					RCODE:   0,
					QDCOUNT: uint16(questionCount),
					ANCOUNT: uint16(len(answers)),
					NSCOUNT: 0,
					ARCOUNT: 0,
				},
				Question: mainQuery.Question,
				Answer:   answers,
			}

			responseBytes := response.Bytes()
			fmt.Printf("Constructed response (%d bytes): %v\n", len(responseBytes), responseBytes)
			responseBytes = stripTrailingZeroValues(responseBytes)

			_, err = udpConn.WriteToUDP(responseBytes, source)
			if err != nil {
				fmt.Println("Failed to send response:", err)
			}
		} else {
			fmt.Printf("No resolver specified, echoing back\n")
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
