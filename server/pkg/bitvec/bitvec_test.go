package bitvec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadBits(t *testing.T) {
	tests := []struct {
		offset int
		data   []byte
		nBits  int
		exp    int
	}{
		{
			data:   []byte{0b0100_0000},
			nBits:  1,
			offset: 1,
			exp:    1,
		},
		{
			data:   []byte{0b0001_0000},
			nBits:  1,
			offset: 3,
			exp:    1,
		},
		{
			data:   []byte{0b0001_1000},
			nBits:  2,
			offset: 3,
			exp:    3,
		},
		{
			data:   []byte{0b0000_1000},
			nBits:  2,
			offset: 4,
			exp:    2,
		},
		{
			data:   []byte{0b0011_1000},
			nBits:  3,
			offset: 2,
			exp:    7,
		},
		{
			data:   []byte{0b0000_1100},
			nBits:  3,
			offset: 3,
			exp:    3,
		},
		{
			data:   []byte{0b0111_0100},
			nBits:  4,
			offset: 1,
			exp:    14,
		},
		{
			data:   []byte{0b0111_0100},
			nBits:  4,
			offset: 2,
			exp:    0b1101,
		},
	}
	for _, tt := range tests {
		vec, err := NewBitVec(tt.data)
		assert.NoError(t, err)
		vec.bitOffset = tt.offset
		res, err := vec.ReadBits(tt.nBits)
		assert.NoError(t, err)
		assert.Equal(t, uint8(tt.exp), res)
	}
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

func TestReadingInvalidNumberOfBits(t *testing.T) {
	data := []byte{0b0}
	vec, err := NewBitVec(data)
	assert.NoError(t, err)
	vec.bitOffset = 6

	res, err := vec.ReadBits(3)
	assert.Error(t, err)
	assert.Equal(t, uint8(0), res)
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
