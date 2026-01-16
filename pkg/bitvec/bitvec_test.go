package bitvec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadBits(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		data   []byte
		nBits  int
		exp    int
	}{
		{
			name:   "single bit at offset 1",
			data:   []byte{0b0100_0000},
			nBits:  1,
			offset: 1,
			exp:    1,
		},
		{
			name:   "Single bit at offset 3",
			data:   []byte{0b0001_0000},
			nBits:  1,
			offset: 3,
			exp:    1,
		},
		{
			name:   "2 bits at offset 3",
			data:   []byte{0b0001_1000},
			nBits:  2,
			offset: 3,
			exp:    3,
		},
		{
			name:   "2 bits at offset 4",
			data:   []byte{0b0000_1000},
			nBits:  2,
			offset: 4,
			exp:    2,
		},
		{
			name:   "3 bits at offset 2",
			data:   []byte{0b0011_1000},
			nBits:  3,
			offset: 2,
			exp:    7,
		},
		{
			name:   "3 bits at offset 3",
			data:   []byte{0b0000_1100},
			nBits:  3,
			offset: 3,
			exp:    3,
		},
		{
			name:   "4 bits at offset 1",
			data:   []byte{0b0111_0100},
			nBits:  4,
			offset: 1,
			exp:    14,
		},
		{
			name:   "4 bits at offset 2",
			data:   []byte{0b0111_0100},
			nBits:  4,
			offset: 2,
			exp:    0b1101,
		},
		{
			name:   "cross byte boundary: 14 bits at offset 2",
			data:   []byte{0b1100_0001, 0b1000_0000},
			nBits:  14,
			offset: 2,
			exp:    0b0000_0001_1000_0000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec, err := NewBitVec(tt.data)
			assert.NoError(t, err)
			vec.pos = tt.offset
			res, err := vec.ReadBits(tt.nBits)
			assert.NoError(t, err)
			assert.Equal(t, uint64(tt.exp), res)
		})
	}
}

func TestPeekBits(t *testing.T) {
	tt := []byte{0b0001_1000}
	vec, err := NewBitVec(tt)
	assert.NoError(t, err)
	vec.bitOffset = 3
	b := vec.PeekBits(2)
	assert.Equal(t, uint8(3), b)
}

func TestReadBytes(t *testing.T) {
	tests := []struct {
		offset int
		data   []byte
		nBytes int
		exp    int
	}{
		{
			data:   []byte{0b0100_0000, 0b0111_0000, 0b1101_1101},
			nBytes: 1,
			offset: 1,
			exp:    0b0111_0000,
		},
		{
			data:   []byte{0b0100_0000, 0b0111_0000, 0b1101_1101},
			nBytes: 2,
			offset: 1,
			exp:    0b0111_0000_1101_1101,
		},
		{
			data:   []byte{0b0100_0000, 0b0111_0000, 0b1101_1101},
			nBytes: 2,
			offset: 0,
			exp:    0b0100_0000_0111_0000,
		},
		{
			data:   []byte{0b0100_0000, 0b0111_0000, 0b1101_1101},
			nBytes: 1,
			offset: 2,
			exp:    0b1101_1101,
		},
		{
			data:   []byte{0b0100_0000, 0b0111_0000, 0b1101_1101, 0b0},
			nBytes: 3,
			offset: 1,
			exp:    0b0111_0000_1101_1101_0000_0000,
		},
	}
	for _, tt := range tests {
		vec, err := NewBitVec(tt.data)
		assert.NoError(t, err)
		vec.byteOffset = tt.offset
		res, err := vec.ReadBytesToUInt32(tt.nBytes)
		assert.NoError(t, err)
		assert.Equal(t, uint32(tt.exp), res)
	}
}

func TestReadingInvalidNumberOfBytes(t *testing.T) {
	data := []byte{0b0100_0000, 0b0111_0000}
	vec, err := NewBitVec(data)
	assert.NoError(t, err)
	vec.byteOffset = 1

	res, err := vec.ReadBytesToUInt32(2)
	assert.Error(t, err)
	assert.Equal(t, uint32(0), res)
}
