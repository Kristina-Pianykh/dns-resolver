package main

import (
	"fmt"
	"net"
	"os"

	"server/pkg/types"
)

type Response struct {
	header Header
	rr     []ResourceRecord
}

type Query struct {
	header   Header
	question Question
}

type Header struct {
	id      types.Byte2
	qr      types.Bit1
	opcode  types.Bit4
	aa      types.Bit1
	tc      types.Bit1
	rd      types.Bit1
	ra      types.Bit1
	z       types.Bit3
	rcode   types.Bit4
	qdcount types.Byte2
	ancount types.Byte2
	nscount types.Byte2
	arcount types.Byte2
}

type Question struct {
	qname  []byte // <N bytes for label>[1 byte]<byte 1>...<byte N>...00000000 (domain name termination)
	qtype  types.Byte2
	qclass types.Byte2
}

type ResourceRecord struct {
	name     []byte
	rrType   types.Byte2
	class    types.Byte2
	ttl      uint32 // seconds
	rdLength types.Byte2
	rdData   byte // variable length, depends on (class, type)
}

func parse(data [512]byte) {
	header := Header{}
	header.id = types.ToByte2(data[0:2])
	header.qr = types.ToBit1(data[0])
	fmt.Printf("query ID: %d\n", int(header.id))
}

func main() {
	// Resolve the string address to a UDP address
	addr := "0.0.0.0:8085"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start listening for UDP packages on the given address
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Read from UDP listener in endless loop
	for {
		var buf [512]byte
		_, _, err := conn.ReadFromUDP(buf[0:])
		fmt.Println("Received a packet")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		parse(buf)

		fmt.Print("> ", string(buf[0:]))

		// Write back the message over UPD
		// conn.WriteToUDP([]byte("Hello UDP Client\n"), addr)
	}
}
