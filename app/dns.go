package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
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
	intValue, err := strconv.Atoi("1")
	if err != nil {
		panic(err)
	}
	questions := make([]Question, 0)
	answers := make([]Answer, 0)

	questionCount := int(binary.BigEndian.Uint16(receivedData[4:6]))
	answerCount := int(binary.BigEndian.Uint16(receivedData[6:8]))

	for i := uint16(0); i < uint16(questionCount); i++ {
		nullB := bytes.Index(receivedData[offset:], []byte{0})
		name, _ := ParseDomain(receivedData, offset)
		dnsType := uint16(intValue)
		class := uint16(intValue)
		dnsTypeBytes := make([]byte, 2)
		classBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(dnsTypeBytes, dnsType)
		binary.BigEndian.PutUint16(classBytes, class)
		questions = append(questions, Question{
			Name:  name,
			Type:  dnsType,
			Class: class,
		})
		offset += nullB + 1
		offset += 4
	}

	for i := 0; i < answerCount; i++ {
		name, nextOffset := ParseDomain(receivedData, offset)
		if nextOffset+10 > len(receivedData) {
			break
		}
		answer := Answer{
			Name:     name,
			Type:     binary.BigEndian.Uint16(receivedData[nextOffset : nextOffset+2]),
			Class:    binary.BigEndian.Uint16(receivedData[nextOffset+2 : nextOffset+4]),
			TTL:      binary.BigEndian.Uint32(receivedData[nextOffset+4 : nextOffset+8]),
			RDLENGTH: binary.BigEndian.Uint16(receivedData[nextOffset+8 : nextOffset+10]),
		}

		offset = nextOffset + 10
		if offset+int(answer.RDLENGTH) > len(receivedData) {
			break
		}
		answer.RDATA = receivedData[offset : offset+int(answer.RDLENGTH)]
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
	originalOffset := offset
	for {
		if offset >= len(data) {
			break
		}
		length := int(data[offset])
		if length == 0 {
			offset++ // Move past the null byte
			break
		}
		if length&0xC0 == 0xC0 { // Check for compression
			if offset+1 >= len(data) {
				return nil, offset + 2 // Return early if not enough data
			}
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			compressedName, _ := ParseDomain(data, pointer)
			name = append(name, compressedName...)
			offset += 2 // Move past the compression pointer
			return name, originalOffset + 2
		}
		name = append(name, data[offset:offset+length+1]...)
		offset += length + 1
	}
	return append(name, 0x00), offset
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
