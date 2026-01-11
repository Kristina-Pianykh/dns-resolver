package parser

import (
	"encoding/hex"
	"fmt"
	"testing"

	"server/pkg/dnsmessage"

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

		fmt.Println(p.Message.Header.String())

		assert.Equal(t, p.Message.Header.ID, tt.expID)
		assert.Equal(t, p.Message.Header.QR, uint64(0))
		assert.Equal(t, p.Message.Header.OpCode, uint64(0))
		assert.Equal(t, p.Message.Header.AA, uint64(0))
		assert.Equal(t, p.Message.Header.TC, uint64(0))
		assert.Equal(t, p.Message.Header.RD, uint64(1))
		assert.Equal(t, p.Message.Header.RA, uint64(0))
		assert.Equal(t, p.Message.Header.Z, uint64(0))
		assert.Equal(t, p.Message.Header.RCode, dnsmessage.RCode(0))
		assert.Equal(t, p.Message.Header.QdCount, uint32(1))
		assert.Equal(t, p.Message.Header.AnCount, uint32(0))
		assert.Equal(t, p.Message.Header.NSCount, uint32(0))
		assert.Equal(t, p.Message.Header.ARCount, uint32(0))

		// p.DebugPrintDomainName(p.Message.Question.QName)
		fmt.Println(p.Message.Question.String())

		assert.Len(t, p.Message.Question.QName, len(tt.expLabels))
		for idx, l := range tt.expLabels {
			assert.Equal(t, []byte(l), p.Message.Question.QName[idx])
		}

		assert.Equal(t, p.Message.Question.QClass, dnsmessage.RRClass(1))
		assert.Equal(t, p.Message.Question.QType, dnsmessage.RRType(1))
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
		assert.NotNil(t, p.Message.Header)
		fmt.Println(p.Message.Header.String())

		assert.Equal(t, p.Message.Header.ID, tt.expID)
		assert.Equal(t, p.Message.Header.QR, uint64(1))
		assert.Equal(t, p.Message.Header.OpCode, uint64(0))
		assert.Equal(t, p.Message.Header.AA, uint64(0))
		assert.Equal(t, p.Message.Header.TC, uint64(0))
		assert.Equal(t, p.Message.Header.RD, uint64(1))
		assert.Equal(t, p.Message.Header.RA, uint64(1))
		assert.Equal(t, p.Message.Header.Z, uint64(0))
		assert.Equal(t, p.Message.Header.RCode, dnsmessage.RCode(0))
		assert.Equal(t, p.Message.Header.QdCount, uint32(1))
		assert.Equal(t, p.Message.Header.AnCount, uint32(1))
		assert.Equal(t, p.Message.Header.NSCount, uint32(0))
		assert.Equal(t, p.Message.Header.ARCount, uint32(0))

		assert.Len(t, p.Message.Question.QName, len(tt.expLabels))
		for idx, l := range tt.expLabels {
			assert.Equal(t, []byte(l), p.Message.Question.QName[idx])
		}

		assert.Equal(t, p.Message.Question.QClass, dnsmessage.RRClass(1))
		assert.Equal(t, p.Message.Question.QType, dnsmessage.RRType(1))
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.vec.SetPos(tt.offset, 0)

			domainName, err := p.ParseLabels(false, -1)
			assert.NoError(t, err)
			fmt.Println(dnsmessage.DomainNameToString(domainName))

			assert.Len(t, domainName, len(tt.expLabels))
		})
	}
}

func TestParseAnswer(t *testing.T) {
	message := "948181800001000500000000086b72697374696e61077069616e796b680378797a0000010001c00c0005000100000635001c106b72697374696e612d7069616e796b680667697468756202696f00c0320001000100000c9b0004b9c76f99c0320001000100000c9b0004b9c76d99c0320001000100000c9b0004b9c76c99c0320001000100000c9b0004b9c76e99"

	expRRs := []dnsmessage.ResourceRecord{
		{
			Name:     dnsmessage.Domain("kristina", "pianykh", "xyz"),
			Type:     dnsmessage.RRType(5),
			Class:    dnsmessage.RRClass(1),
			TTL:      uint32(1589),
			RdLength: uint32(28),
			RData:    []byte("kristina-pianykhgithubio"),
		},
		{
			Name:     dnsmessage.Domain("kristina-pianykh", "github", "io"),
			Type:     dnsmessage.RRType(1),
			Class:    dnsmessage.RRClass(1),
			TTL:      uint32(3227),
			RdLength: uint32(4),
			RData:    []byte{185, 199, 111, 153},
		},
		{
			Name:     dnsmessage.Domain("kristina-pianykh", "github", "io"),
			Type:     dnsmessage.RRType(1),
			Class:    dnsmessage.RRClass(1),
			TTL:      uint32(3227),
			RdLength: uint32(4),
			RData:    []byte{185, 199, 109, 153},
		},
		{
			Name:     dnsmessage.Domain("kristina-pianykh", "github", "io"),
			Type:     dnsmessage.RRType(1),
			Class:    dnsmessage.RRClass(1),
			TTL:      uint32(3227),
			RdLength: uint32(4),
			RData:    []byte{185, 199, 108, 153},
		},
		{
			Name:     dnsmessage.Domain("kristina-pianykh", "github", "io"),
			Type:     dnsmessage.RRType(1),
			Class:    dnsmessage.RRClass(1),
			TTL:      uint32(3227),
			RdLength: uint32(4),
			RData:    []byte{185, 199, 110, 153},
		},
	}

	input, err := hex.DecodeString(message)
	assert.NoError(t, err)

	arr := [512]byte{}
	copy(arr[:], input[:])

	p, err := NewParser(arr)
	assert.NoError(t, err)

	err = p.ParseHeader()
	assert.NoError(t, err)

	err = p.ParseQuestion()
	assert.NoError(t, err)

	err = p.ParseAnswer()
	assert.NoError(t, err)

	fmt.Println(p.Message.String())

	assert.Len(t, p.Message.Answers, 5)

	for i := range expRRs {
		expRR := expRRs[i]
		actRR := *p.Message.Answers[i]
		assert.Equal(t, expRR, actRR)
	}
}

func TestParseMessage(t *testing.T) {
	message := "04068000000100000004000806676f6f676c6503636f6d0000020001c00c000200010002a3000006036e7332c00cc00c000200010002a3000006036e7331c00cc00c000200010002a3000006036e7333c00cc00c000200010002a3000006036e7334c00cc028001c00010002a30000102001486048020034000000000000000ac028000100010002a3000004d8ef220ac03a001c00010002a30000102001486048020032000000000000000ac03a000100010002a3000004d8ef200ac04c001c00010002a30000102001486048020036000000000000000ac04c000100010002a3000004d8ef240ac05e001c00010002a30000102001486048020038000000000000000ac05e000100010002a3000004d8ef260a"
	input, err := hex.DecodeString(message)
	assert.NoError(t, err)

	arr := [512]byte{}
	copy(arr[:], input[:])

	p, err := NewParser(arr)
	assert.NoError(t, err)

	err = p.ParseMessage()
	assert.NoError(t, err)

	fmt.Println(p.Message)
	assert.Len(t, p.Message.Answers, 0)
	assert.Len(t, p.Message.AuthorityRecords, 4)
	assert.Len(t, p.Message.AdditonalRecords, 8)
}
