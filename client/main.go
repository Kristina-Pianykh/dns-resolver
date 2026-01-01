package main

import (
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"sync"
)

func main() {
	addr := ":8085"
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rInt := rand.IntN(10)
	var wg sync.WaitGroup
	for i := range rInt {
		wg.Go(func() {
			sendUDPMessage(i, conn)
		})
	}

	wg.Wait()
}

func sendUDPMessage(id int, conn *net.UDPConn) {
	rInt := rand.IntN(10)
	for i := range rInt {
		message := fmt.Sprintf("ID: %d, message: %d\n", id, i)

		// Send a message to the server
		_, err := conn.Write([]byte(message))
		fmt.Printf("send %s", message)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
