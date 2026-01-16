package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"server/pkg/log"
	"server/pkg/parser"
)

type ServerConfig struct {
	UDPCfg  UDPConfig
	TCPCfg  TCPConfig
	Timeout time.Duration
}

type UDPConfig struct {
	Addr          string
	Port          int
	MaxBufferSize int
	Timeout       time.Duration
}

type TCPConfig struct {
	Addr    string
	Port    int
	Timeout time.Duration
}

type Server struct {
	Cfg       *ServerConfig
	servers   []NetworkServer
	UDPServer *UDPServer
	// TCPServer *TCPServer
}

type NetworkServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	GetNet() string
}

func NewServer() (*Server, error) {
	srvCfg := ServerConfig{
		UDPCfg: UDPConfig{
			Addr:          "",
			Port:          8085,
			MaxBufferSize: 512,
		},
		Timeout: 5 * time.Second,
	}

	servers := []NetworkServer{}
	udpSrv, err := NewUDPServer(&srvCfg.UDPCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP server: %w", err)
	}
	servers = append(servers, udpSrv)

	srv := Server{
		Cfg:     &srvCfg,
		servers: servers,
	}
	return &srv, nil
}

func (s *Server) Start(ctx context.Context, errChan chan error) {
	log.Info("starting server...")

	go func() {
		for _, server := range s.servers {
			log.Info("starting %s server...", server.GetNet())
			if err := server.Start(ctx); err != nil {
				errChan <- fmt.Errorf("start %s listener failed: %w", server.GetNet(), err)
				return
			}
		}
	}()
}

// Stop stops the server
func (s *Server) Stop(ctx context.Context) error {
	log.Info("Stopping server")

	for _, server := range s.servers {
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("stop %s listener failed: %w", server.GetNet(), err)
		}
	}

	return nil
}

type UDPServer struct {
	Config *UDPConfig
	Conn   *net.UDPConn
	Net    string
}

func (s *UDPServer) GetNet() string {
	return s.Net
}

func NewUDPServer(cfg *UDPConfig) (*UDPServer, error) {
	addr := net.UDPAddr{
		Port: cfg.Port,
		IP:   net.ParseIP(cfg.Addr),
	}

	UDPConn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP listener: %w", err)
	}

	srv := UDPServer{
		Config: cfg,
		Conn:   UDPConn,
		Net:    "udp",
	}

	return &srv, nil
}

func (s *UDPServer) Start(ctx context.Context) error {
	log.Info("starting UDP server on %s:%d...", s.Config.Addr, s.Config.Port)
	if s.Conn == nil {
		return errors.New("UDP connection is not initialized")
	}

	// if parent context is done, close the UDP connection
	go func() {
		<-ctx.Done()
		log.Debug("context done, closing UDP connection...")
		_ = s.Conn.Close()
	}()

	for {
		buf := make([]byte, s.Config.MaxBufferSize)
		n, addr, err := s.Conn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Debug("UDP server got error: %s", err.Error())

			if errors.Is(err, net.ErrClosed) {
				log.Info("UDP connection closed, exiting read loop")
				return nil
			} else {
				log.Debug("error type: %T", err)
			}

			select {
			case <-ctx.Done():
				log.Info("UDP server shutting down")
			default:
				log.Error("error reading from UDP: %w", err)
				continue
			}
		}
		log.Info("Received a packet (%d bytes) from %s", n, addr.String())

		// TODO: set a timeout for the session
		go func(data []byte, addr *net.UDPAddr) {
			timeoutCtx, cancel := context.WithTimeout(ctx, s.Config.Timeout)
			defer cancel()

			DNSProcess(data, addr, s.Conn, timeoutCtx)
		}(buf, addr)
	}
}

func (s *UDPServer) Shutdown(ctx context.Context) error {
	log.Info("shutting down UDP server...")
	if s.Conn != nil {
		log.Debug("closing UDP connection...")
		if err := s.Conn.Close(); err != nil {
			return fmt.Errorf("failed to close UDP connection: %w", err)
		}
	}
	return nil
}

func DNSProcess(data []byte, addr *net.UDPAddr, conn *net.UDPConn, ctx context.Context) {
	log.Info("Processing DNS data from %s", addr.String())

	// w := ResponseWriter{
	// 	Conn: conn,
	// 	Addr: addr,
	// }

	p, err := parser.NewParser(data)
	if err != nil {
		// TODO: handle error
		log.Error("failed to create DNS parser: %s", err)
		return
	}

	err = p.ParseMessage()
	if err != nil {
		log.Error("failed to parse DNS message: %v", err)
		return
	}

	if p.Message.IsQuery() {
		_, err := conn.WriteToUDP([]byte("received a query. Placeholder for resolving"), addr)
		if err != nil {
			log.Error("failed to write DNS query response to", addr.String(), ":", err)
			return
		}
		// resolve query
		// make a dns query
		// block, wait for response or timeout
		// } else {
		// 	// handle response
		// 	// write dns in binary as udpo back to addr
		// 	_, err := conn.WriteToUDP(data, addr)
		// 	if err != nil {
		// 		log.Error("failed to write DNS response to", addr.String(), ":", err)
		// 		return
		// 	}
	}

	// Here you would parse the DNS message and respond accordingly
	log.Info("Finished processing DNS data from %s", addr.String())
}

type ResponseWriter struct {
	Conn *net.UDPConn
	Addr *net.UDPAddr
	data []byte
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.Conn.WriteToUDP(data, rw.Addr)
	if err != nil {
		return n, fmt.Errorf("failed to write data to %s: %w", rw.Addr.String(), err)
	}
	return n, nil
}
