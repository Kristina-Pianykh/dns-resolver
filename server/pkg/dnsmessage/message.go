package dnsmessage

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

type DNSMessage struct {
	Header   *Header
	Question *Question
	Answer   *Answer
}

type Answer struct{}

type Question struct {
	QName  [][]byte // <N bytes for label>[1 byte]<byte 1>...<byte N>...00000000 (domain name termination)
	QType  uint32
	QClass uint32
}

type ResourceRecord struct {
	name     []byte
	rrType   uint32
	class    uint32
	ttl      uint32 // seconds
	rdLength uint32
	rdData   byte // variable length, depends on (class, type)
}
