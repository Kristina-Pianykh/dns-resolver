package dnsmessage

import (
	"fmt"
	"strings"

	"server/pkg/log"
)

type (
	Label      = []byte
	DomainName = []Label
)

type (
	RRType          uint32
	RRClass         uint32
	RCode           uint64
	ResourceRecords []*ResourceRecord
)

const (
	TypeA     RRType = 1
	TypeNS    RRType = 2  // Authoritative name server
	TypeMD    RRType = 3  // Mail destination (obsolete)
	TypeMF    RRType = 4  // Mail forwarder (obsolete)
	TypeCNAME RRType = 5  // Canonical name
	TypeSOA   RRType = 6  // Start of authority
	TypeMB    RRType = 7  // Mailbox domain name (experimental)
	TypeMG    RRType = 8  // Mail group member (experimental)
	TypeMR    RRType = 9  // Mail rename domain name (experimental)
	TypeNULL  RRType = 10 // Null RDATA (ignore)
	TypeWKS   RRType = 11 // Well-known service (deprecated)
	TypePTR   RRType = 12 // Pointer to a domain name
	TypeHINFO RRType = 13 // Host info
	TypeMINFO RRType = 14 // Mailbox or mail list info (experimental)
	TypeMX    RRType = 15 // Mail exchange (currently rejected)
	TypeTXT   RRType = 16 // Text strings

	TypeAAAA RRType = 28 // IPv6
)

const (
	ClassIN RRClass = 1
	ClassCH RRClass = 3
	ClassHS RRClass = 4
)

type Header struct {
	ID      uint32
	QR      uint64
	OpCode  uint64
	AA      uint64
	TC      uint64
	RD      uint64
	RA      uint64
	Z       uint64
	RCode   RCode
	QdCount uint32
	AnCount uint32
	NSCount uint32
	ARCount uint32
}

type DNSMessage struct {
	Header           *Header
	Question         *Question
	Answers          ResourceRecords
	AuthorityRecords ResourceRecords
	AdditonalRecords ResourceRecords
}

type Question struct {
	QName  [][]byte // <N bytes for label>[1 byte]<byte 1>...<byte N>...00000000 (domain name termination)
	QType  RRType
	QClass RRClass
}

type ResourceRecord struct {
	Name     DomainName
	Type     RRType
	Class    RRClass
	TTL      uint32 // seconds
	RdLength uint32
	RData    []byte // variable length, depends on (class, type)
}

func (m *DNSMessage) IsQuery() bool {
	return m.Header.QR == 0
}

func (m *DNSMessage) String() string {
	if m == nil {
		log.Debug("DNSMessage is nil")
		return ""
	}
	return fmt.Sprintf(`
%s

%s

ANSWERS: %s

AUTHORITY: %s

ADDITIONAL RECORDS: %s`, m.Header, m.Question, m.Answers, m.AuthorityRecords, m.AdditonalRecords)
}

func (h *Header) String() string {
	return fmt.Sprintf(`HEADER:
  ID: %d
  QR: %s
  OpCode: %s
  AA (Authoritative Answer): %d
  TC (Truncation): %d
  RD (Recursion Desired): %d
  RA (Recursion Available): %d
  Z: %d
  RCode: %s
  Question Count: %d
  Answer Count: %d
  NS Resource Records: %d
  RRs in additional records section: %d`, h.ID,
		QRToString(h.QR),
		OpCodeToString(h.OpCode),
		h.AA,
		h.TC,
		h.RD,
		h.RA,
		h.Z,
		h.RCode,
		h.QdCount,
		h.AnCount,
		h.NSCount,
		h.ARCount)
}

func (rr ResourceRecords) String() string {
	if len(rr) == 0 {
		log.Debug("empty or nil Resource Records")
		return ""
	}
	var sb strings.Builder
	for _, r := range rr {
		sb.WriteString(r.String())
	}
	return sb.String()
}

func (c RCode) String() string {
	switch int(c) {
	case 0:
		return "0 - no error"
	case 1:
		// The name server was unable to interpret the query.
		return "1 - format error"
	case 2:
		// The name server was unable to process this query due to a problem with the name server.
		return "2 - server failure"
	case 3:
		// Meaningful only for responses from an authoritative name
		// server, this code signifies that the domain name
		// referenced in the query does not exist.
		return "3 - name error"
	case 4:
		// The name server does not support the requested kind of query.
		return "4 - not implemented"
	case 5:
		return "5 - refused"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(c))
	}
}

func (q *Question) String() string {
	return fmt.Sprintf(`QUESTION:
  Qname: %s
  QType: %s
  QClass: %s`, DomainNameToString(q.QName), q.QType, q.QClass)
}

func (rr *ResourceRecord) String() string {
	if rr == nil {
		log.Debug("Resource Record is nil")
		return ""
	}

	var rdata string

	switch rr.Type {
	case TypeA, TypeAAAA:
		for i, b := range rr.RData {
			rdata += fmt.Sprintf("%d", b)
			if i < len(rr.RData)-1 {
				rdata += "."
			}
		}
	case TypeCNAME, TypeNS:
		for _, b := range rr.RData {
			rdata += string(b)
		}
		// TODO: implement the rest
	}

	return fmt.Sprintf(`
Resource Record:
  Name: %s
  Type: %s
  Class: %s
  TTL: %d sec
  RData Length: %d
  RData: `, DomainNameToString(rr.Name), rr.Type, rr.Class, rr.TTL, rr.RdLength) + rdata
}

func DomainNameToString(name DomainName) string {
	if len(name) == 0 {
		log.Debug("Domain Name is nil or empty")
		return ""
	}
	var sb strings.Builder
	for i, label := range name {
		sb.WriteString(string(label))
		if i < len(name)-1 {
			sb.WriteString(".")
		}
	}
	return sb.String()
}

func QRToString(v uint64) string {
	switch int(v) {
	case 0:
		return "query"
	case 1:
		return "response"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(v))
	}
}

func OpCodeToString(c uint64) string {
	switch int(c) {
	case 0:
		return "standard query"
	case 1:
		return "inverse query"
	case 2:
		return "server status request"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(c))
	}
}

func (c RRClass) String() string {
	switch c {
	case ClassIN:
		return "IN"
	case ClassCH:
		return "CH"
	case ClassHS:
		return "HS"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", c)
	}
}

func (t RRType) String() string {
	switch t {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeMD:
		return "MD"
	case TypeMF:
		return "MF"
	case TypeCNAME:
		return "CNAME"
	case TypeSOA:
		return "SOA"
	case TypeMB:
		return "MB"
	case TypeMG:
		return "MG"
	case TypeMR:
		return "MR"
	case TypeNULL:
		return "NULL"
	case TypeWKS:
		return "WKS"
	case TypePTR:
		return "PTR"
	case TypeHINFO:
		return "HINFO"
	case TypeMINFO:
		return "MINFO"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

func Domain(labels ...string) DomainName {
	dn := make(DomainName, len(labels))
	for i, l := range labels {
		dn[i] = []byte(l)
	}
	return dn
}
