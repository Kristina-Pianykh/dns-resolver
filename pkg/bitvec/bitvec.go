package bitvec

import (
	"errors"
	"fmt"
	"io"
)

const MaxLength = 512

type BitVec struct {
	data       []byte
	pos        int
	byteOffset int
	bitOffset  int
}

func (v *BitVec) GetPos() (int, int) {
	return v.byteOffset, v.bitOffset
}

func (v *BitVec) SetPos(byteOffset int, bitOffset int) {
	v.byteOffset = byteOffset
	v.bitOffset = bitOffset
	v.pos = v.byteOffset*8 + bitOffset
}

func NewBitVec(data []byte) (BitVec, error) {
	if len(data) > MaxLength {
		return BitVec{}, fmt.Errorf("input data is bigger than %d bytes", MaxLength)
	}
	d := make([]byte, 0, 512)
	copy(d, data)
	return BitVec{data: data}, nil
}

func (v *BitVec) updateOffsets(bitInc int) {
	v.pos += bitInc
	v.bitOffset = v.pos % 8
	v.byteOffset = v.pos / 8
}

func (v *BitVec) PeekBits(n int) uint8 {
	shift := 8 - v.bitOffset - n
	shifted := v.data[v.byteOffset] >> shift
	return shifted & (255 >> (8 - n))
}

func (v *BitVec) ReadBits(n int) (uint64, error) {
	if n <= 0 {
		return 0, fmt.Errorf("invalid bit count: %d", n)
	}

	if n > 64 {
		return 0, fmt.Errorf("cannot read more than 64 bits at once: %d", n)
	}

	if v.pos+n > len(v.data)*8 {
		return 0, io.ErrUnexpectedEOF
	}

	var result uint64

	for i := 0; i < n; i++ {
		byteIdx := v.pos / 8
		shift := 7 - (v.pos % 8)

		bit := (v.data[byteIdx] >> shift) & 1
		result = (result << 1) | uint64(bit)

		v.pos++
	}

	v.bitOffset = v.pos % 8
	v.byteOffset = v.pos / 8

	return result, nil
}

func (v *BitVec) ReadBytesToUInt32(n int) (uint32, error) {
	arr, err := v.ReadBytes(n)
	if err != nil {
		return 0, fmt.Errorf("failed to read %d bytes: %w", n, err)
	}

	return BytesToUint32(arr), nil
}

func (v *BitVec) ReadBytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("can't read %d bytes", n)
	}

	if v.byteOffset >= 512 {
		return nil, errors.New("reached the end of the packet")
	}

	if n+v.byteOffset > MaxLength || n+v.byteOffset > len(v.data) {
		return nil, fmt.Errorf("read out of bounds for reading %d bytes at pos %d", n, v.pos)
	}

	// byte-long fields align with byte boundaries so bitOffset
	// cannot be > 0
	if v.bitOffset != 0 {
		return nil, fmt.Errorf("violated byte boundery when reading %d bytes", n)
	}
	arr := v.data[v.byteOffset : v.byteOffset+n]
	v.updateOffsets(8 * n)

	return arr, nil
}

func BytesToUint32(b []byte) uint32 {
	var val uint32
	n := len(b)

	for i := range n {
		shift := (n - i - 1) * 8
		// log.Debug("i: %d, shift: %d\n", i, shift)
		// log.Debug("uint32(b[i])<<shift: %08b\n", uint32(b[i])<<shift)
		val = val | uint32(b[i])<<shift
	}
	// log.Debug("val: %08b\n", val)
	return val
}
