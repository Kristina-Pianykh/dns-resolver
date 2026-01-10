package parser

import (
	"errors"
	"fmt"

	"server/pkg/bitvec"
	"server/pkg/dnsmessage"
	message "server/pkg/dnsmessage"
	"server/pkg/log"
)

const MaxLabelLength = 63 // in bytes

type Parser struct {
	vec     *bitvec.BitVec
	Header  *message.Header
	Message *message.DNSMessage
}

func NewParser(data [512]byte) (Parser, error) {
	vec, err := bitvec.NewBitVec(data[:])
	if err != nil {
		return Parser{}, fmt.Errorf("failed to initialize Parser: %w\n", err)
	}
	dnsMessage := message.DNSMessage{}
	return Parser{vec: &vec, Message: &dnsMessage}, nil
}

// fields never cross byte bounderies so we don't care about alignment

func (p *Parser) ParseQuestion() error {
	// we assume there's only one question
	question := message.Question{}

	domainName, err := p.ParseLabels(false, -1)
	if err != nil {
		return fmt.Errorf("failed to parse labels: %w", err)
	}

	question.QName = domainName
	qType, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return fmt.Errorf("failed to parse QType from Question: %w", err)
	}
	question.QType = qType
	log.Debug("QType: %d", qType)

	qClass, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return fmt.Errorf("failed to parse QClass from Question: %w", err)
	}
	log.Debug("QClass: %d", qClass)
	question.QClass = qClass

	p.Message.Question = &question

	return nil
}

func (p *Parser) ParseLabels(rec bool, byteOffset int) ([][]byte, error) {
	// TODO: handle stack overflow from too many recursions
	// TODO: handle potential infinite loops from malicious pointers
	var (
		origByteOffset int
		origBitOffset  int
	)

	if rec {
		if byteOffset < 0 {
			return nil, errors.New("byte offset as a pointer for reading a label must be >= 0")
		}
		origByteOffset, origBitOffset = p.vec.GetPos()
		p.vec.SetPos(int(byteOffset), 0)
	}

	defer func() {
		if rec {
			p.vec.SetPos(origByteOffset, origBitOffset)
		}
	}()

	labels := [][]byte{}

	lengthByte, err := p.vec.ReadBytesToUInt32(1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length byte: %w", err)
	}

	for lengthByte > 0 {
		// compression
		if lengthByte&0xC0 == 0xC0 {
			log.Debug("recursive label parsing")
			b2, err := p.vec.ReadBytesToUInt32(1)
			if err != nil {
				return nil, err
			}

			v := (lengthByte << 8) | b2
			ptr := v & 0x3fff
			log.Debug("ptr: %d", ptr)

			l, err := p.ParseLabels(true, int(ptr))
			if err != nil {
				return l, err
			}

			labels = append(labels, l...)
			return labels, nil
		}

		if lengthByte > uint32(MaxLabelLength) {
			return nil, fmt.Errorf("a label can't be bigger than %d bytes", MaxLabelLength)
		}

		// normal label
		label, err := p.vec.ReadBytes(int(lengthByte))
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)

		lengthByte, err = p.vec.ReadBytesToUInt32(1)
		if err != nil {
			return nil, fmt.Errorf("failed to parse length byte: %w", err)
		}

	}
	return labels, nil
}

func ByteArrToASCII(bytes []byte) string {
	s := ""
	for _, b := range bytes {
		s += string(b)
	}
	return s
}

func (p *Parser) ParseMessage() error {
	err := p.ParseHeader()
	if err != nil {
		return fmt.Errorf("failed to parse headers: %v", err)
	}

	err = p.ParseQuestion()
	if err != nil {
		return fmt.Errorf("failed to parse Question: %v", err)
	}
	return nil
}

func (p *Parser) ParseAnswer() error {
	resourceRecords := []dnsmessage.ResourceRecord{}

	rr := dnsmessage.ResourceRecord{}
	domainName, err := p.ParseLabels(false, -1)
	if err != nil {
		return fmt.Errorf("failed to parse resource record labels: %w", err)
	}
	rr.Name = domainName

	rrType, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return fmt.Errorf("failed to parse resource record type: %w", err)
	}
	rr.Type = rrType

	class, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return fmt.Errorf("failed to parse resource record class: %w", err)
	}
	rr.Class = class

	ttl, err := p.vec.ReadBytesToUInt32(4)
	if err != nil {
		return fmt.Errorf("failed to parse resource record ttl: %w", err)
	}
	rr.TTL = ttl

	rdLength, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return fmt.Errorf("failed to parse length of additional data for resource record: %w", err)
	}
	rr.RdLength = rdLength

	// TODO: parse RData [page 12]

	resourceRecords = append(resourceRecords, rr)
	return nil
}

func (p *Parser) parseRData(rType uint32) ([]byte, error) {
	switch t := int(rType); t {
	// A
	case 1:
		log.Debug("parsing type A RDATA")
		v, err := p.vec.ReadBytes(4)
		if err != nil {
			return nil, fmt.Errorf("failed to parse type A RDATA: %w", err)
		}
		return v, nil

	// NS
	case 2:
		// this should never reach our server
		log.Debug("parsing NS RDATA")
		v, err := p.ParseLabels(false, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to parse domain name in NS RDATA: %w", err)
		}
		flattened := []byte{}
		for _, label := range v {
			flattened = append(flattened, label...)
		}
		return flattened, nil

	// MD (3) - mail destination - obsolete
	// MF (4) - mail forwarder - obsolete
	// SOA (6) - marks the start of a zone of authority (not sure if we need this now)
	// MB (7) - mailbox domain name - experimental (we choose to reject)
	// MG (8) - main group member - experimental (we choose to reject)
	// MR (9) - main rename domain name - experimental (we choose to reject)
	// MINFO (14) - mailbox or mail list info - experimental (we choose to reject)
	// MX (15) - mail exchange (we choose to reject for now)
	case 3, 4, 6, 7, 8, 9, 14, 15:
	// TODO: unimplemented

	// CNAME
	case 5:
		log.Debug("parsing CNAME RDATA")
		v, err := p.ParseLabels(false, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to parse domain name in CNAME RDATA: %w", err)
		}
		flattened := []byte{}
		for _, label := range v {
			flattened = append(flattened, label...)
		}
		return flattened, nil

	// NULL RDATA - ignore
	case 10:
		return []byte{}, nil

	// WKS - a well known service description
	case 11:
	// PTR
	case 12:
		log.Debug("parsing PTR RDATA")
		v, err := p.ParseLabels(false, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to parse domain name in PTR RDATA: %w", err)
		}
		flattened := []byte{}
		for _, label := range v {
			flattened = append(flattened, label...)
		}
		return flattened, nil

	// HINFO - host info
	case 13:
		return nil, fmt.Errorf("unknown RR TYPE: %d", t)
		// reject
	}

	// unreachable
	return nil, nil
}

func (p *Parser) ParseHeader() error {
	header := dnsmessage.Header{}

	id, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return err
	}
	header.ID = id

	qr, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.QR = qr

	opCode, err := p.vec.ReadBits(4)
	if err != nil {
		return err
	}
	header.OpCode = opCode

	aa, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.AA = aa

	tc, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.TC = tc

	rd, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.RD = rd

	ra, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.RA = ra

	z, err := p.vec.ReadBits(3)
	if err != nil {
		return err
	}
	header.Z = z

	rCode, err := p.vec.ReadBits(4)
	if err != nil {
		return err
	}
	header.RCode = rCode

	qdCount, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return err
	}
	header.QdCount = qdCount

	anCount, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return err
	}
	header.AnCount = anCount

	nsCount, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return err
	}
	header.NSCount = nsCount

	arCount, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return err
	}
	header.ARCount = arCount

	p.Header = &header
	return nil
}

func (p *Parser) DebugPrintHeader() {
	log.Debug("query ID: %d\n", int(p.Header.ID))
	log.Debug("QR: %d\n", int(p.Header.QR))
	log.Debug("Opcode: %d\n", int(p.Header.OpCode))
	log.Debug("AA: %d\n", int(p.Header.AA))
	log.Debug("TC: %d\n", int(p.Header.TC))
	log.Debug("RD: %d\n", int(p.Header.RD))
	log.Debug("RA: %d\n", int(p.Header.RA))
	log.Debug("Z: %d\n", int(p.Header.Z))
	log.Debug("Rcode: %d\n", int(p.Header.RCode))
	log.Debug("QDCount: %d\n", int(p.Header.QdCount))
	log.Debug("ANCount: %d\n", int(p.Header.AnCount))
	log.Debug("NSCount: %d\n", int(p.Header.NSCount))
	log.Debug("ARCount: %d\n", int(p.Header.ARCount))
}

func (p *Parser) DebugPrintDomainName(name dnsmessage.DomainName) {
	for i, label := range name {
		log.Debug("Label %d: %s\n", i, label)
	}
}
