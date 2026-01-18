package server

import (
	"fmt"
	"net"
)

type ResponseWriter struct {
	Conn *net.UDPConn
	Addr *net.UDPAddr
}

func NewResponseWriter(conn *net.UDPConn, addr string) (*ResponseWriter, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	return &ResponseWriter{Conn: conn, Addr: udpAddr}, nil
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.Conn.WriteToUDP(data, rw.Addr)
	if err != nil {
		return n, fmt.Errorf("failed to write data to %s: %w", rw.Addr.String(), err)
	}
	return n, nil
}
