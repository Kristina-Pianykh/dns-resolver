package bitvec

import (
	"errors"
	"fmt"

	"server/pkg/log"
)

const MaxLength = 512

type BitVec struct {
	data       []byte
	pos        int
	bitOffset  int
	byteOffset int
}

func NewBitVec(data []byte) (BitVec, error) {
	if len(data) > MaxLength {
		return BitVec{}, fmt.Errorf("input data is bigger than %d bytes", MaxLength)
	}
	d := make([]byte, 0, 512)
	copy(d, data)
	return BitVec{data: data}, nil
}

func (v *BitVec) updateOffsets() {
	v.bitOffset = v.pos % 8
	v.byteOffset = v.pos / 8
}

// 0b0100_0000
func (v *BitVec) ReadBits(n int) (uint8, error) {
	if n < 0 || n > 7 {
		return 0, fmt.Errorf("Bit offset into byte must be between 0 and 7, got %d", n)
	}

	if v.bitOffset+n > 8 {
		return 0, fmt.Errorf("violated byte boundary when reading %d bits", n)
	}
	shift := 8 - v.bitOffset - n             // 6
	shifted := v.data[v.byteOffset] >> shift // 1 in decimal
	log.Debug("shift: %d\n", shift)
	log.Debug("shifted: %08b\n", shifted)
	log.Debug("255 >> (8 - %d): %b\n", n, 255>>(8-n))
	masked := shifted & (255 >> (8 - n))
	log.Debug("masked: %08b\n", masked)
	v.pos += n
	v.updateOffsets()
	return masked, nil
}

func (v *BitVec) ReadBytesToInt(n int) (uint32, error) {
	if n < 0 {
		return 0, fmt.Errorf("Can't read %d bytes", n)
	}

	if v.byteOffset >= 512 {
		// TODO: signal nothing to do, reach end of the packet
		return 0, errors.New("Reached the end of the packet")
	}

	if n+v.byteOffset > MaxLength || n+v.byteOffset > len(v.data) {
		return 0, fmt.Errorf("Can't read %d + %d > %d", v.byteOffset, n, MaxLength)
	}

	// byte-long fields align with byte boundaries so bitOffset
	// cannot be > 0
	if v.bitOffset != 0 {
		return 0, errors.New("Violated byte boundery when reading 2 bytes. Exiting.")
	}
	b := v.data[v.byteOffset : v.byteOffset+n]
	v.pos += 8 * n
	v.updateOffsets()

	// Examples
	// b[0]<<8 | b[1]
	// b[0]<<24 | b[1]<<16 | b[2]<<8 | b[3]

	var val uint32
	for i := 0; i < n; i++ {
		shift := (n - i - 1) * 8
		log.Debug("i: %d, shift: %d\n", i, shift)
		log.Debug("uint32(b[i])<<shift: %08b\n", uint32(b[i])<<shift)
		val = val | uint32(b[i])<<shift
	}
	log.Debug("val: %08b\n", val)
	return val, nil
}

func (v *BitVec) ReadBytesToArr(n int) ([]byte, error) {
	// 510, 511
	if v.byteOffset+n > MaxLength {
		return nil, fmt.Errorf("Read out of bounds: %d + %d > %d", v.byteOffset, n, MaxLength)
	}

	arr := v.data[v.byteOffset : v.byteOffset+n]
	v.pos += 8 * n
	v.updateOffsets()
	return arr, nil
}
