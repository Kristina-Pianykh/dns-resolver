package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"server/pkg/parser"
)

func parse(data [512]byte) error {
	p, err := parser.NewParser(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = p.ParseMessage()
	if err != nil {
		return fmt.Errorf("failed to parse message: %v", err)
	}
	p.DebugPrintHeader()
	p.DebugPrintQueryLabels()

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
