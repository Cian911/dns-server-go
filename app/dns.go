package main

import (
	"encoding/binary"
)

type Message struct {
	Header         Header
	Question       []Question
	ResourceRecord []ResourceRecord
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

type ResourceRecord struct {
	Name     []byte // A domain name, represented as a suqeuence of "labels"
	Type     uint16 // 2-byte int; the type of record (1 for A record, 5 for CNAME)
	Class    uint16 // 2-byte int; usually set to 1, for "IN"
	TTL      uint32 // The duration in seconds a record can be cached before requerying.
	RDLENGTH uint16 // Length of the RDATA field
	RDATA    []byte // Data specific to the record type, such as an IPv4 address
}

func NewQuery() *Message {
	return &Message{
		Header: Header{
			ID:      1234,
			QR:      1,
			OPCODE:  0,
			AA:      0,
			TC:      0,
			RD:      0,
			RA:      0,
			Z:       0,
			RCODE:   0,
			QDCOUNT: 1,
			ANCOUNT: 1,
			NSCOUNT: 0,
			ARCOUNT: 0,
		},
		Question: []Question{
			{
				Name:  []byte("\x0ccodecrafters\x02io\x00"),
				Type:  1,
				Class: 1,
			},
		},
		ResourceRecord: []ResourceRecord{
			{
				Name:     []byte("\x0ccodecrafters\x02io\x00"),
				Type:     1,
				Class:    1,
				TTL:      60,
				RDLENGTH: 4,
				RDATA:    []byte("\x08\x08\x08\x08"),
			},
		},
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
