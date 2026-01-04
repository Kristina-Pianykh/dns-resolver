package dnsmessage

import "server/pkg/types"

type Response struct {
	header Header
	rr     []ResourceRecord
}

type Query struct {
	header   Header
	question Question
}

type Header struct {
	ID      uint32
	QR      uint8
	OpCode  uint8
	AA      uint8
	TC      uint8
	RD      uint8
	RA      uint8
	Z       uint8
	RCode   uint8
	QdCount uint32
	AnCount uint32
	NSCount uint32
	ARCount uint32
}

type Question struct {
	qname  []byte // <N bytes for label>[1 byte]<byte 1>...<byte N>...00000000 (domain name termination)
	qtype  types.Byte2
	qclass types.Byte2
}

type ResourceRecord struct {
	name     []byte
	rrType   types.Byte2
	class    types.Byte2
	ttl      uint32 // seconds
	rdLength types.Byte2
	rdData   byte // variable length, depends on (class, type)
}
