package parser

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		data      string
		expLabels []string
		expID     uint32
	}{
		{
			data:      "7e4e01000001000000000000076e69636b6c6173077365646c6f636b0378797a0000010001",
			expLabels: []string{"nicklas", "sedlock", "xyz"},
			expID:     32334,
		},
		{
			data:      "45dc010000010000000000000377777707796f757475626503636f6d0000010001",
			expLabels: []string{"www", "youtube", "com"},
			expID:     17884,
		},
	}
	for _, tt := range tests {
		input, err := hex.DecodeString(tt.data)
		assert.NoError(t, err)

		arr := [512]byte{}
		copy(arr[:], input[:])

		p, err := NewParser(arr)
		assert.NoError(t, err)

		err = p.ParseMessage()
		assert.NoError(t, err)
		p.DebugPrintHeader()
		assert.Equal(t, p.Header.ID, tt.expID)
		assert.Equal(t, p.Header.QR, uint8(0))
		assert.Equal(t, p.Header.OpCode, uint8(0))
		assert.Equal(t, p.Header.AA, uint8(0))
		assert.Equal(t, p.Header.TC, uint8(0))
		assert.Equal(t, p.Header.RD, uint8(1))
		assert.Equal(t, p.Header.RA, uint8(0))
		assert.Equal(t, p.Header.Z, uint8(0))
		assert.Equal(t, p.Header.RCode, uint8(0))
		assert.Equal(t, p.Header.QdCount, uint32(1))
		assert.Equal(t, p.Header.AnCount, uint32(0))
		assert.Equal(t, p.Header.NSCount, uint32(0))
		assert.Equal(t, p.Header.ARCount, uint32(0))

		p.DebugPrintQueryLabels()
		assert.Len(t, p.Message.Question.QName, len(tt.expLabels))
		for idx, l := range tt.expLabels {
			assert.Equal(t, []byte(l), p.Message.Question.QName[idx])
		}

		assert.Equal(t, p.Message.Question.QClass, uint32(1))
		assert.Equal(t, p.Message.Question.QType, uint32(1))
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		data      string
		expID     uint32
		expLabels []string
	}{
		{
			data:      "deb1818000010001000000000377777706676f6f676c6503636f6d0000010001c00c000100010000001300048efabaa4",
			expID:     57009,
			expLabels: []string{"www", "google", "com"},
		},
	}
	for _, tt := range tests {
		input, err := hex.DecodeString(tt.data)
		assert.NoError(t, err)

		arr := [512]byte{}
		copy(arr[:], input[:])

		p, err := NewParser(arr)
		assert.NoError(t, err)

		err = p.ParseMessage()
		assert.NoError(t, err)
		p.DebugPrintHeader()

		assert.Equal(t, p.Header.ID, tt.expID)
		assert.Equal(t, p.Header.QR, uint8(1))
		assert.Equal(t, p.Header.OpCode, uint8(0))
		assert.Equal(t, p.Header.AA, uint8(0))
		assert.Equal(t, p.Header.TC, uint8(0))
		assert.Equal(t, p.Header.RD, uint8(1))
		assert.Equal(t, p.Header.RA, uint8(1))
		assert.Equal(t, p.Header.Z, uint8(0))
		assert.Equal(t, p.Header.RCode, uint8(0))
		assert.Equal(t, p.Header.QdCount, uint32(1))
		assert.Equal(t, p.Header.AnCount, uint32(1))
		assert.Equal(t, p.Header.NSCount, uint32(0))
		assert.Equal(t, p.Header.ARCount, uint32(0))

		assert.Len(t, p.Message.Question.QName, len(tt.expLabels))
		for idx, l := range tt.expLabels {
			assert.Equal(t, []byte(l), p.Message.Question.QName[idx])
		}

		assert.Equal(t, p.Message.Question.QClass, uint32(1))
		assert.Equal(t, p.Message.Question.QType, uint32(1))
	}
}
