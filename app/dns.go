package main

import (
	"encoding/binary"
	"strings"
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
	questions := make([]Question, 0)
	answers := make([]Answer, 0)
	for i := 0; i < int(binary.BigEndian.Uint16(receivedData[4:6])); i++ {
		answer := Answer{
			Name:     ParseDomain(receivedData),
			Type:     1,
			Class:    1,
			TTL:      60,
			RDLENGTH: 4,
			RDATA:    []byte("\x08\x08\x08\x08"),
		}
		answers = append(answers, answer)
	}
	for i := 0; i < int(binary.BigEndian.Uint16(receivedData[4:6])); i++ {
		question := Question{
			Name:  ParseDomain(receivedData),
			Type:  1,
			Class: 1,
		}

		questions = append(questions, question)
	}

	return &Message{
		Header: Header{
			ID:      binary.BigEndian.Uint16(receivedData[0:2]),
			QR:      1,
			OPCODE:  uint8(binary.BigEndian.Uint16(receivedData[2:4]) >> 11),
			AA:      0,
			TC:      0,
			RD:      uint8(binary.BigEndian.Uint16(receivedData[2:4]) >> 8),
			RA:      0,
			Z:       0,
			RCODE:   4,
			QDCOUNT: uint16(binary.BigEndian.Uint16(receivedData[4:6])),
			ANCOUNT: uint16(binary.BigEndian.Uint16(receivedData[4:6])),
			NSCOUNT: 0,
			ARCOUNT: 0,
		},
		Question: questions,
		ResourceRecord: []ResourceRecord{
			{
				Name:     ParseDomain(receivedData),
				Type:     1,
				Class:    1,
				TTL:      60,
				RDLENGTH: 4,
				RDATA:    []byte("\x08\x08\x08\x08"),
			},
		},
		Answer: answers,
	}
}

func (m *Message) Bytes() []byte {
	response := m.Header.Bytes()
	for _, question := range m.Question {
		response = append(response, question.Bytes()...)
	}
	for _, rr := range m.ResourceRecord {
		response = append(response, rr.Bytes()...)
	}
	for _, a := range m.Answer {
		response = append(response, a.Bytes()...)
	}

	return response
}

func (h *Header) Bytes() []byte {
	// Header Message is 12 bytes long
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], h.ID)
	flag := uint16(h.QR)<<15 |
		uint16(h.OPCODE)<<11 |
		uint16(h.AA)<<10 |
		uint16(h.TC)<<9 |
		uint16(h.RD)<<8 |
		uint16(h.RA)<<7 |
		uint16(h.Z)<<4 |
		uint16(h.RCODE)
	binary.BigEndian.PutUint16(buf[2:4], flag)
	binary.BigEndian.PutUint16(buf[4:6], h.QDCOUNT)
	binary.BigEndian.PutUint16(buf[6:8], h.ANCOUNT)
	binary.BigEndian.PutUint16(buf[8:10], h.NSCOUNT)
	binary.BigEndian.PutUint16(buf[10:12], h.ARCOUNT)

	return buf
}

func (q *Question) Bytes() []byte {
	t := make([]byte, 2)
	binary.BigEndian.PutUint16(t, q.Type)
	c := make([]byte, 2)
	binary.BigEndian.PutUint16(c, q.Class)

	return append(append(q.Name, t...), c...)
}

func (rr *ResourceRecord) Bytes() []byte {
	var buf []byte
	t := make([]byte, 2)
	binary.BigEndian.PutUint16(t, rr.Type)
	c := make([]byte, 2)
	binary.BigEndian.PutUint16(c, rr.Class)
	ttl := make([]byte, 4)
	binary.BigEndian.PutUint32(ttl, rr.TTL)
	l := make([]byte, 2)
	binary.BigEndian.PutUint16(l, rr.RDLENGTH)

	buf = append(buf, rr.Name...)
	buf = append(buf, t...)
	buf = append(buf, c...)
	buf = append(buf, ttl...)
	buf = append(buf, l...)
	buf = append(buf, rr.RDATA...)

	return buf
}

func (a *Answer) Bytes() []byte {
	var buf []byte
	t := make([]byte, 2)
	binary.BigEndian.PutUint16(t, a.Type)
	c := make([]byte, 2)
	binary.BigEndian.PutUint16(c, a.Class)
	ttl := make([]byte, 4)
	binary.BigEndian.PutUint32(ttl, a.TTL)
	l := make([]byte, 2)
	binary.BigEndian.PutUint16(l, a.RDLENGTH)

	buf = append(buf, a.Name...)
	buf = append(buf, t...)
	buf = append(buf, c...)
	buf = append(buf, ttl...)
	buf = append(buf, l...)
	buf = append(buf, a.RDATA...)

	return buf
}

func ParseDomain(data []byte) []byte {
	domainByte := data[12:]
	domain := decodeDNSPacket(domainByte)

	segments := strings.Split(domain, ".")
	var encodedDomain []byte
	for _, segment := range segments {
		encodedDomain = append(encodedDomain, byte(len(segment)))
		encodedDomain = append(encodedDomain, []byte(segment)...)
	}

	// Terminate with null byte
	encodedDomain = append(encodedDomain, 0x00)
	return encodedDomain
}

func decodeDNSPacket(packet []byte) string {
	var domain string
	i := 0
	for i < len(packet) && packet[i] != 0 {
		labelLength := int(packet[i])
		i++
		if i+labelLength > len(packet) {
			break
		}
		domain += string(packet[i : i+labelLength])
		i += labelLength
		if i < len(packet) && packet[i] != 0 {
			domain += "."
		}
	}
	return domain
}
