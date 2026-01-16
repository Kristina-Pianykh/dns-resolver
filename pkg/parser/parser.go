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
	Message *message.DNSMessage
}

func NewParser(data []byte) (Parser, error) {
	vec, err := bitvec.NewBitVec(data)
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
	question.QType = dnsmessage.RRType(qType)
	log.Debug("QType: %d", qType)

	qClass, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return fmt.Errorf("failed to parse QClass from Question: %w", err)
	}
	log.Debug("QClass: %d", qClass)
	question.QClass = dnsmessage.RRClass(qClass)

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

	err = p.ParseAnswer()
	if err != nil {
		return fmt.Errorf("failed to parse Answer: %w", err)
	}

	err = p.ParseAuthority()
	if err != nil {
		return fmt.Errorf("failed to parse Authority: %w", err)
	}

	err = p.ParseAdditionalRRs()
	if err != nil {
		return fmt.Errorf("failed to parse Additional Recourds: %w", err)
	}
	return nil
}

func (p *Parser) ParseAnswer() error {
	resourceRecords := []*dnsmessage.ResourceRecord{}
	log.Debug("Answers: %d", p.Message.Header.AnCount)

	for i := range p.Message.Header.AnCount {
		rr, err := p.parseRR()
		if err != nil {
			return fmt.Errorf("failed to parse a resource record at idx %d (answer section): %w", i, err)
		}
		resourceRecords = append(resourceRecords, rr)
	}

	p.Message.Answers = resourceRecords
	return nil
}

func (p *Parser) ParseAuthority() error {
	resourceRecords := []*dnsmessage.ResourceRecord{}
	log.Debug("Authority Records: %d", p.Message.Header.NSCount)

	for i := range p.Message.Header.NSCount {
		rr, err := p.parseRR()
		if err != nil {
			return fmt.Errorf("failed to parse a resource record at idx %d (authority section): %w", i, err)
		}
		resourceRecords = append(resourceRecords, rr)
	}

	p.Message.AuthorityRecords = resourceRecords
	return nil
}

func (p *Parser) ParseAdditionalRRs() error {
	resourceRecords := []*dnsmessage.ResourceRecord{}
	log.Debug("Authority Records: %d", p.Message.Header.ARCount)

	for i := range p.Message.Header.ARCount {
		rr, err := p.parseRR()
		if err != nil {
			return fmt.Errorf("failed to parse a resource record at idx %d (additional section): %w", i, err)
		}
		resourceRecords = append(resourceRecords, rr)
	}

	p.Message.AdditonalRecords = resourceRecords
	return nil
}

func (p *Parser) parseRR() (*dnsmessage.ResourceRecord, error) {
	rr := dnsmessage.ResourceRecord{}
	domainName, err := p.ParseLabels(false, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource record labels: %w", err)
	}
	rr.Name = domainName

	rrType, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource record type: %w", err)
	}
	rr.Type = dnsmessage.RRType(rrType)

	class, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource record class: %w", err)
	}
	rr.Class = dnsmessage.RRClass(class)

	ttl, err := p.vec.ReadBytesToUInt32(4)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource record ttl: %w", err)
	}
	rr.TTL = ttl

	rdLength, err := p.vec.ReadBytesToUInt32(2)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length of additional data for resource record: %w", err)
	}
	rr.RdLength = rdLength

	// TODO: parse RData [page 12]
	rData, err := p.parseRData(rr.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RDATA: %w", err)
	}
	rr.RData = rData
	return &rr, nil
}

func (p *Parser) parseRData(rType dnsmessage.RRType) ([]byte, error) {
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
		return nil, fmt.Errorf("unimplemented %d type", t)

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
		// TODO
		return nil, fmt.Errorf("unimplemented %d type", t)

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
	case 28:
		log.Debug("parsing type AAAA RDATA")
		v, err := p.vec.ReadBytes(16)
		if err != nil {
			return nil, fmt.Errorf("failed to parse type AAAA RDATA: %w", err)
		}
		return v, nil
		// reject
	default:
		return nil, fmt.Errorf("UNKNOWN(%d)", t)
	}
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
	header.RCode = dnsmessage.RCode(rCode)

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

	p.Message.Header = &header
	return nil
}
