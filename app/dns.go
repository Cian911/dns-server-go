package main

import (
	"encoding/binary"
	"fmt"
)

type Message struct {
	Header         Header
	Question       []Question
	ResourceRecord []ResourceRecord
	Answer         []Answer
}

/*
The header section is 12 bytes long. Ints are encoded in big-endian.
*/
type Header struct {
	ID      uint16 // Packet Identifier
	QR      uint8  // Query/Response Indicator (1 bit)
	OPCODE  uint8  // Operation code (4 bits)
	AA      uint8  // Authoritative answer (1 bit)
	TC      uint8  // Truncation (1 bit)
	RD      uint8  // Recusion desired (1 bit)
	RA      uint8  // Recusion available (1 bit)
	Z       uint8  // Used by DNSSEC queries. (3 bits)
	RCODE   uint8  // Response code (4 bits)
	QDCOUNT uint16 // Question count
	ANCOUNT uint16 // Answer record count
	NSCOUNT uint16 // Autority record count
	ARCOUNT uint16 // Additional record count
}

type Question struct {
	Name  []byte // A domain name, represented as a suqeuence of "labels"
	Type  uint16 // 2-byte int; the type of record (1 for A record, 5 for CNAME)
	Class uint16 // 2-byte int; usually set to 1, for "IN"
}

type Answer struct {
	Name     []byte // A domain name, represented as a suqeuence of "labels"
	Type     uint16 // 2-byte int; the type of record (1 for A record, 5 for CNAME)
	Class    uint16 // 2-byte int; usually set to 1, for "IN"
	TTL      uint32 // The duration in seconds a record can be cached before requerying.
	RDLENGTH uint16 // Length of the RDATA field
	RDATA    []byte // Data specific to the record type, such as an IPv4 address
}

type ResourceRecord struct {
	Name     []byte // A domain name, represented as a suqeuence of "labels"
	Type     uint16 // 2-byte int; the type of record (1 for A record, 5 for CNAME)
	Class    uint16 // 2-byte int; usually set to 1, for "IN"
	TTL      uint32 // The duration in seconds a record can be cached before requerying.
	RDLENGTH uint16 // Length of the RDATA field
	RDATA    []byte // Data specific to the record type, such as an IPv4 address
}

func NewQuery(receivedData []byte) *Message {
	offset := 12
	questions := make([]Question, 0)
	answers := make([]Answer, 0)

	questionCount := int(binary.BigEndian.Uint16(receivedData[4:6]))
	answerCount := int(binary.BigEndian.Uint16(receivedData[6:8]))

	for i := 0; i < questionCount; i++ {
		name, nextOffset := ParseDomain(receivedData, offset)
		offset = nextOffset + 4
		question := Question{
			Name:  name,
			Type:  binary.BigEndian.Uint16(receivedData[nextOffset : nextOffset+2]),
			Class: binary.BigEndian.Uint16(receivedData[nextOffset+2 : nextOffset+4]),
		}
		questions = append(questions, question)
	}

	for i := 0; i < answerCount; i++ {
		name, nextOffset := ParseDomain(receivedData, offset)
		offset = nextOffset
		answer := Answer{
			Name:     name,
			Type:     binary.BigEndian.Uint16(receivedData[offset : offset+2]),
			Class:    binary.BigEndian.Uint16(receivedData[offset+2 : offset+4]),
			TTL:      binary.BigEndian.Uint32(receivedData[offset+4 : offset+8]),
			RDLENGTH: binary.BigEndian.Uint16(receivedData[offset+8 : offset+10]),
		}

		fmt.Printf("Name: %s, Type: %d, Class: %d, TTL: %d, RDLENGTH: %d\n", formatDomainName(answer.Name), answer.Type, answer.Class, answer.TTL, answer.RDLENGTH)

		offset += 10

		// if offset+int(answer.RDLENGTH) > len(receivedData) {
		// 	log.Fatalf("RDLENGTH too large, cannot slice: offset=%d, RDLENGTH=%d, len(receivedData)=%d", offset, answer.RDLENGTH, len(receivedData))
		// }

		answer.RDATA = receivedData[offset : offset+int(answer.RDLENGTH)]
		fmt.Printf("RDATA: %v\n", answer.RDATA)

		offset += int(answer.RDLENGTH)
		answers = append(answers, answer)
	}

	return &Message{
		Header: Header{
			ID:      binary.BigEndian.Uint16(receivedData[0:2]),
			QR:      (receivedData[2] >> 7) & 0x1,
			OPCODE:  (receivedData[2] >> 3) & 0xf,
			AA:      (receivedData[2] >> 2) & 0x1,
			TC:      (receivedData[2] >> 1) & 0x1,
			RD:      receivedData[2] & 0x1,
			RA:      (receivedData[3] >> 7) & 0x1,
			Z:       (receivedData[3] >> 4) & 0x7,
			RCODE:   receivedData[3] & 0xf,
			QDCOUNT: binary.BigEndian.Uint16(receivedData[4:6]),
			ANCOUNT: binary.BigEndian.Uint16(receivedData[6:8]),
			NSCOUNT: binary.BigEndian.Uint16(receivedData[8:10]),
			ARCOUNT: binary.BigEndian.Uint16(receivedData[10:12]),
		},
		Question: questions,
		Answer:   answers,
	}
}

func (m *Message) Bytes() []byte {
	response := m.Header.Bytes()
	for _, question := range m.Question {
		response = append(response, question.Bytes()...)
	}
	for _, a := range m.Answer {
		response = append(response, a.Bytes()...)
	}
	return response
}

func (h *Header) Bytes() []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], h.ID)
	buf[2] = (h.QR << 7) | (h.OPCODE << 3) | (h.AA << 2) | (h.TC << 1) | h.RD
	buf[3] = (h.RA << 7) | (h.Z << 4) | h.RCODE
	binary.BigEndian.PutUint16(buf[4:6], h.QDCOUNT)
	binary.BigEndian.PutUint16(buf[6:8], h.ANCOUNT)
	binary.BigEndian.PutUint16(buf[8:10], h.NSCOUNT)
	binary.BigEndian.PutUint16(buf[10:12], h.ARCOUNT)
	return buf
}

func (q *Question) Bytes() []byte {
	buf := append([]byte{}, q.Name...)
	t := make([]byte, 2)
	binary.BigEndian.PutUint16(t, q.Type)
	buf = append(buf, t...)
	c := make([]byte, 2)
	binary.BigEndian.PutUint16(c, q.Class)
	return append(buf, c...)
}

func (a *Answer) Bytes() []byte {
	buf := append([]byte{}, a.Name...)
	t := make([]byte, 2)
	binary.BigEndian.PutUint16(t, a.Type)
	buf = append(buf, t...)
	c := make([]byte, 2)
	binary.BigEndian.PutUint16(c, a.Class)
	buf = append(buf, c...)
	ttl := make([]byte, 4)
	binary.BigEndian.PutUint32(ttl, a.TTL)
	buf = append(buf, ttl...)
	l := make([]byte, 2)
	binary.BigEndian.PutUint16(l, a.RDLENGTH)
	buf = append(buf, l...)
	return append(buf, a.RDATA...)
}

func ParseDomain(data []byte, offset int) ([]byte, int) {
	var name []byte
	for {
		length := int(data[offset])
		if length == 0 {
			break
		}
		/*
		   Compression Handling: The ParseDomain function is updated to check if the length byte has its two highest bits set (length&0xC0 == 0xC0). If so, it handles the name as a pointer to another part of the packet, using the lower 14 bits of the next two bytes as the new offset.
		*/
		if length&0xC0 == 0xC0 { // Check for compression
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			compressedName, _ := ParseDomain(data, pointer)
			name = append(name, compressedName...)
			offset += 2
			break
		}
		name = append(name, data[offset:offset+length+1]...)
		offset += length + 1
	}
	return append(name, 0x00), offset + 1
}

func formatDomainName(name []byte) string {
	var result string
	for i := 0; i < len(name); {
		length := int(name[i])
		if length == 0 {
			break
		}
		if i != 0 {
			result += "."
		}
		result += string(name[i+1 : i+1+length])
		i += length + 1
	}
	return result
}
