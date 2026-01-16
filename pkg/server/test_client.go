package server

import "net"

type UDPClient struct {
	Conn *net.UDPConn
	Name string
}

func NewUDPClient(serverAddr string) (*UDPClient, error) {
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

	_, err := c.Conn.Write(data)
	if err != nil {
		return nil, err
	}

	n, _, err := c.Conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:n], nil
}
