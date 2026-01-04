package parser

import (
	"fmt"

	"server/pkg/bitvec"
	"server/pkg/dnsmessage"
	message "server/pkg/dnsmessage"
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

func (p *Parser) ParseHeader() error {
	header := dnsmessage.Header{}

	// log.Debug("Reading ID at pos: %d, byteOffset: %d, bitOffset: %d\n", p.vec.pos, p.byteOffset, p.bitOffset)
	id, err := p.vec.ReadBytes(2)
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
	qdCount, err := p.vec.ReadBytes(2)
	if err != nil {
		return err
	}
	header.QdCount = qdCount

	// log.Debug("Reading AnCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	anCount, err := p.vec.ReadBytes(2)
	if err != nil {
		return err
	}
	header.AnCount = anCount

	// log.Debug("Reading NSCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	nsCount, err := p.vec.ReadBytes(2)
	if err != nil {
		return err
	}
	header.NSCount = nsCount

	// log.Debug("Reading ARCount at pos: %d, byteOffset: %d, bitOffset: %d\n", p.pos, p.byteOffset, p.bitOffset)
	arCount, err := p.vec.ReadBytes(2)
	if err != nil {
		return err
	}
	header.ARCount = arCount

	p.Header = &header
	return nil
}
