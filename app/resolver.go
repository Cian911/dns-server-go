package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// Forward request to resolver address
func Forward(addr string, packet []byte, size int) {
	// Create buffer
	buffer := make([]byte, 1024)

	for {
		// Listen from connection
		ip, port := parseAddr(addr)
		forwardAddr := net.UDPAddr{
			Port: port,
			IP:   net.ParseIP(ip),
		}
		forwardConn, err := net.DialUDP("udp", nil, &forwardAddr)
		if err != nil {
			log.Fatalf("Could not dial remote server: %v", err)
		}
		defer forwardConn.Close()
		// Write to connection
		_, err = forwardConn.Write(packet[:size])
		if err != nil {
			log.Fatalf("Error writing data to server: %v", err)
		}
		fmt.Printf("Forwarded %d to %v\n", size, forwardAddr)
		// Read from connection
		forwardConn.SetReadDeadline(time.Now().Add(3 * time.Second))
		size, source, err := forwardConn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalf("Error reading from remote server: %v", err)
		}

		fmt.Printf("Received %d bytes from %s: %s\n", size, source, string(buffer))
	}
}

func parseAddr(addr string) (string, int) {
	strSplit := strings.Split(addr, ":")
	if len(strSplit) == 1 {
		return strSplit[0], 0
	}
	port, err := strconv.Atoi(strSplit[1])
	if err != nil {
		log.Fatal(err)
	}
	return strSplit[0], port
}
