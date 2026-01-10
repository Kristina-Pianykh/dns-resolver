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
		assert.Equal(t, p.Header.QR, uint64(0))
		assert.Equal(t, p.Header.OpCode, uint64(0))
		assert.Equal(t, p.Header.AA, uint64(0))
		assert.Equal(t, p.Header.TC, uint64(0))
		assert.Equal(t, p.Header.RD, uint64(1))
		assert.Equal(t, p.Header.RA, uint64(0))
		assert.Equal(t, p.Header.Z, uint64(0))
		assert.Equal(t, p.Header.RCode, uint64(0))
		assert.Equal(t, p.Header.QdCount, uint32(1))
		assert.Equal(t, p.Header.AnCount, uint32(0))
		assert.Equal(t, p.Header.NSCount, uint32(0))
		assert.Equal(t, p.Header.ARCount, uint32(0))

		p.DebugPrintDomainName(p.Message.Question.QName)
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
		assert.Equal(t, p.Header.QR, uint64(1))
		assert.Equal(t, p.Header.OpCode, uint64(0))
		assert.Equal(t, p.Header.AA, uint64(0))
		assert.Equal(t, p.Header.TC, uint64(0))
		assert.Equal(t, p.Header.RD, uint64(1))
		assert.Equal(t, p.Header.RA, uint64(1))
		assert.Equal(t, p.Header.Z, uint64(0))
		assert.Equal(t, p.Header.RCode, uint64(0))
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

func TestParsingLabels(t *testing.T) {
	// Test compression format for labels from RFC 1035
	// F.ISI.ARPA, FOO.F.ISI.ARPA, ARPA, and the root
	//
	// byte 0:  length 1, F
	// byte 2:  length 3, I
	// byte 4:         S, I
	// byte 6:  length 4, A
	// byte 8:         R, P
	// byte 10:        A, 0
	// byte 12: length 3, F
	// byte 14:        O, O
	// byte 16: 11   20
	// byte 18: 11   06
	// byte 20:      00
	names := "01460349534904415250410003464F4FC000C00600"
	tests := []struct {
		name      string
		offset    int
		expLabels []string
	}{
		{
			name:      "offset 0",
			offset:    0,
			expLabels: []string{"F", "ISI", "ARPA"},
		},
		{
			name:      "offset 12",
			offset:    12,
			expLabels: []string{"FOO", "F", "ISI", "ARPA"},
		},
		{
			name:      "offset 18",
			offset:    18,
			expLabels: []string{"ARPA"},
		},
	}
	input, err := hex.DecodeString(names)
	assert.NoError(t, err)

	arr := [512]byte{}
	copy(arr[:], input[:])

	p, err := NewParser(arr)
	assert.NoError(t, err)

	// we skip the first 20 bytes of headers
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.vec.SetPos(tt.offset, 0)

			domainName, err := p.ParseLabels(false, -1)
			assert.NoError(t, err)
			p.DebugPrintDomainName(domainName)

			assert.Len(t, domainName, len(tt.expLabels))
		})
	}
}
