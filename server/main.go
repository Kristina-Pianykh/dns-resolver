package main

import (
	"fmt"
	"net"
	"os"
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
	id      uint16
	qr      uint8 // 1 bit
	opcode  uint8 // 4 bits
	aa      uint8 // 1 bit
	tc      uint8 // 1 bit
	rd      uint8 // 1 bit
	ra      uint8 // 1 bit
	z       uint8 // 3 bits
	rcode   uint8 // 4 bits
	qdcount uint16
	ancount uint16
	nscount uint16
	arcount uint16
}

type Question struct {
	qname  []byte // <N bytes for label>[1 byte]<byte 1>...<byte N>...00000000 (domain name termination)
	qtype  uint16
	qclass uint16
}

type ResourceRecord struct {
	name     []byte
	rrType   uint16
	class    uint16
	ttl      uint32 // seconds
	rdLength uint16
	rdData   byte // variable length, depends on (class, type)
}

func parse(data [512]byte) {
	header := Header{}
	header.id = uint16(data[0])<<8 | uint16(data[1])
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
