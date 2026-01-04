package main

import (
	"fmt"
	"log"
	"net"
	"os"

	logger "server/pkg/log"
	"server/pkg/parser"
)

func parse(data [512]byte) error {
	p, err := parser.NewParser(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	err = p.ParseHeader()
	if err != nil {
		return fmt.Errorf("Failed to parse headers: %v", err)
	}

	logger.Debug("query ID: %d\n", int(p.Header.ID))
	logger.Debug("QR: %d\n", int(p.Header.QR))
	logger.Debug("Opcode: %d\n", int(p.Header.OpCode))
	logger.Debug("AA: %d\n", int(p.Header.AA))
	logger.Debug("TC: %d\n", int(p.Header.TC))
	logger.Debug("RD: %d\n", int(p.Header.RD))
	logger.Debug("RA: %d\n", int(p.Header.RA))
	logger.Debug("Z: %d\n", int(p.Header.Z))
	logger.Debug("Rcode: %d\n", int(p.Header.RCode))
	logger.Debug("QDCount: %d\n", int(p.Header.QdCount))
	logger.Debug("ANCount: %d\n", int(p.Header.AnCount))
	logger.Debug("NSCount: %d\n", int(p.Header.NSCount))
	logger.Debug("ARCount: %d\n", int(p.Header.ARCount))

	err = p.ParseQuestion()
	if err != nil {
		fmt.Errorf("Failed to parse Question: %v", err)
	}
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags)

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

		err = parse(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse: %v", err)
			os.Exit(1)
		}

		fmt.Print("> ", string(buf[0:]))

		// Write back the message over UPD
		// conn.WriteToUDP([]byte("Hello UDP Client\n"), addr)
	}
}
