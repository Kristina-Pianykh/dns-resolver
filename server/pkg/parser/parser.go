package parser

import (
	"fmt"

	"server/pkg/bitvec"
	"server/pkg/dnsmessage"
	message "server/pkg/dnsmessage"
	"server/pkg/log"
)

type Parser struct {
	vec      *bitvec.BitVec
	Header   *message.Header
	Question *message.Question
}

// 0, 8, 16, ...
// if pos mod 8 == 0 --> slice into byte array.
// else:
// e.g. pos = 2 -> byte[0] << 2 || (byte[1] && 0b

func NewParser(data [512]byte) (Parser, error) {
	vec, err := bitvec.NewBitVec(data[:])
	if err != nil {
		return Parser{}, fmt.Errorf("failed to initialize Parser: %w\n", err)
	}
	return Parser{vec: &vec}, nil
}

// fields never cross byte bounderies so we don't care about alignment

func (p *Parser) ParseQuestion() error {
	// we assume there's only one question
	question := message.Question{}

	labels, err := p.ParseLabels()
	if err != nil {
		return fmt.Errorf("failed to parse labels: %w", err)
	}

	for i, label := range labels {
		log.Debug("Label %d: %s\n", i, label)
	}

	question.QName = labels
	qType, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return fmt.Errorf("failed to parse QType from Question: %w", err)
	}
	question.QType = qType
	log.Debug("QType: %d", qType)

	qClass, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return fmt.Errorf("failed to parse QClass from Question: %w", err)
	}
	log.Debug("QClass: %d", qClass)
	question.QClass = qClass
	p.Question = &question

	return nil
}

func (p *Parser) ParseLabels() ([][]byte, error) {
	labels := [][]byte{}
	var length uint32
	var err error

	length, err = p.vec.ReadBytesToInt(1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length byte: %w", err)
	}

	for length > 0 {
		if err != nil {
			return nil, err
		}
		label, err := p.vec.ReadBytesToArr(int(length))
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
		length, err = p.vec.ReadBytesToInt(1)
		if err != nil {
			return nil, fmt.Errorf("failed to parse length byte: %w", err)
		}
	}

	return labels, nil
}

func ByteArrToAscii(bytes []byte) string {
	// var sb strings.Builder
	s := ""
	for _, b := range bytes {
		s += string(b)
	}
	return s
}

func (p *Parser) ParseHeader() error {
	header := dnsmessage.Header{}

	// log.Debug("Reading ID at pos: %d, byteOffset: %d, bitOffset: %d\n", p.vec.pos, p.byteOffset, p.bitOffset)
	id, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return err
	}
	header.ID = id

	// log.Debug("Reading QR at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	qr, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.QR = qr

	// log.Debug("Reading OpCode at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	opCode, err := p.vec.ReadBits(4)
	if err != nil {
		return err
	}
	header.OpCode = opCode

	// log.Debug("Reading AA at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	aa, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.AA = aa

	// log.Debug("Reading TC at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	tc, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.TC = tc

	// log.Debug("Reading RD at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	rd, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.RD = rd

	// log.Debug("Reading RA at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	ra, err := p.vec.ReadBits(1)
	if err != nil {
		return err
	}
	header.RA = ra

	// log.Debug("Reading Z at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	z, err := p.vec.ReadBits(3)
	if err != nil {
		return err
	}
	header.Z = z

	// log.Debug("Reading RCode at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	rCode, err := p.vec.ReadBits(4)
	if err != nil {
		return err
	}
	header.RCode = rCode

	// log.Debug("Reading QdCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	qdCount, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return err
	}
	header.QdCount = qdCount

	// log.Debug("Reading AnCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	anCount, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return err
	}
	header.AnCount = anCount

	// log.Debug("Reading NSCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	nsCount, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return err
	}
	header.NSCount = nsCount

	// log.Debug("Reading ARCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	arCount, err := p.vec.ReadBytesToInt(2)
	if err != nil {
		return err
	}
	header.ARCount = arCount

	p.Header = &header
	return nil
}
