package main

import (
	"math/rand"
	"time"
)


type Message struct {
  header Header
}

/*
  The header section is 12 bytes long. Ints are encoded in big-endian.
*/
type Header struct {
  ID byte // Packet Identifier
  QR byte // Query/Response Indicator
  OPCODE byte // Operation code
  AA byte // Authoritative answer
  TC byte // Truncation
  RD byte // Recusion desired
  RA byte // Recusion available
  Z byte // Used by DNSSEC queries. 
  RCODE byte // Response code
  QDCOUNT byte // Question count
  ANCOUNT byte // Answer record count
  NSCOUNT byte // Autority record count
  ARCOUNT byte // Additional record count
}

func NewQuery(data []byte) *Message {
  
  return &Message{
    header: Header{
      ID: byte(generatePacketId()),
      QR: byte(1),
      OPCODE: byte(0),
      AA: byte(0),
    },
  }
}

func generatePacketId() uint8 {
  return uint8(rand.Intn(1 << 8))
}
