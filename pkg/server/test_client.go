package server

import (
	"fmt"
	"net"
	"time"
)

type UDPClient struct {
	Conn    *net.UDPConn
	Name    string
	Timeout time.Duration
}

func NewUDPClient(serverAddr string, timeout time.Duration) (*UDPClient, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	client := UDPClient{
		Conn: conn,
	}

	return &client, nil
}

func (c *UDPClient) Send(data []byte) (int, error) {
	return c.Conn.Write(data)
}

func (c *UDPClient) SendAndReceive(data []byte, bufferSize int) ([]byte, error) {
	buffer := make([]byte, bufferSize)

	// fail fast instead of hanging forever if no response arrives
	_ = c.Conn.SetReadDeadline(time.Now().Add(c.Timeout))

	_, err := c.Conn.Write(data)
	if err != nil {
		return nil, err
	}
	fmt.Printf("sent %s\n", string(data))

	n, _, err := c.Conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	fmt.Printf("read %d bytes\n", n)

	return buffer[:n], nil
}

func (c *UDPClient) Close() error {
	return c.Conn.Close()
}
