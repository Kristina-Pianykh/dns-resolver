package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToBit1(t *testing.T) {
	bytes := []byte{0b1000_0000, 0b0100_0000}

	bit := ToBit1(bytes)
	assert.Equal(t, 1, int(bit))
	assert.Equal(t, 0, int(bytes[0]))
	assert.Equal(t, 128, int(bytes[1]))
}

func TestToBit2(t *testing.T) {
	bytes := []byte{0b1011_0000, 0b010_0000}

	bit := ToBit2(bytes)
	assert.Equal(t, 2, int(bit))
	assert.Equal(t, 192, int(bytes[0]))
	assert.Equal(t, 128, int(bytes[1]))
}

func TestToBit3(t *testing.T) {
	bytes := []byte{0b1011_0000, 0b1010_0001}

	bit := ToBit3(bytes)
	assert.Equal(t, 5, int(bit))
	assert.Equal(t, 128, int(bytes[0]))
	assert.Equal(t, 8, int(bytes[1]))
}

func TestToBit4(t *testing.T) {
	bytes := []byte{0b1110_1000, 0b1010_0001}

	bit := ToBit4(bytes)
	assert.Equal(t, 14, int(bit))
	assert.Equal(t, 128, int(bytes[0]))
	assert.Equal(t, 16, int(bytes[1]))
}

func TestToByte2(t *testing.T) {
	bytes := []byte{0b1110_0101, 0b0110_1001}
	target := ToByte2(&bytes)
	assert.Equal(t, 58729, int(target))
	assert.Empty(t, bytes)
}
