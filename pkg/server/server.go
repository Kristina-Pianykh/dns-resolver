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
	Start(ctx context.Context, errChan chan error) error
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

func (s *Server) Start(ctx context.Context, errChan chan error, procErrChan chan error) {
	log.Info("starting server...")

	go func() {
		for _, server := range s.servers {
			log.Info("starting %s server...", server.GetNet())
			if err := server.Start(ctx, procErrChan); err != nil {
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

func (s *UDPServer) Start(ctx context.Context, errChan chan error) error {
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
		buf := make([]byte, s.Config.MaxBufferSize*3)
		n, addr, err := s.Conn.ReadFromUDP(buf[0:])

		// test it
		if n > s.Config.MaxBufferSize {
			errChan <- fmt.Errorf("received packet size %d exceeds max buffer size %d", n, s.Config.MaxBufferSize)
			continue
		}

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
				return nil
			default:
				log.Warn("error reading from UDP: %w", err)
				continue
			}
		}
		log.Info("Received a packet (%d bytes) from %s", n, addr.String())

		// TODO: set a timeout for the session
		go func(data []byte, addr *net.UDPAddr) {
			timeoutCtx, cancel := context.WithTimeout(ctx, s.Config.Timeout)
			defer cancel()

			DNSProcess(data, addr, s.Conn, timeoutCtx, errChan)
		}(buf[:n], addr)
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

func DNSProcess(data []byte, addr *net.UDPAddr, conn *net.UDPConn, ctx context.Context, errChan chan error) {
	log.Info("Processing DNS data from %s", addr.String())

	p, err := parser.NewParser(data)
	if err != nil {
		// TODO: handle error
		errChan <- fmt.Errorf("failed to create DNS parser: %w", err)
		return
	}

	err = p.ParseMessage()
	if err != nil {
		errChan <- fmt.Errorf("failed to parse DNS message: %v", err)
		return
	}
	m := p.Message

	if m.IsQuery() {
		// handle dns query
		log.Info("Processing DNS query")

		TransactionTable.Store(int(m.Header.ID), Addr{Addr: addr.String(), Timestamp: time.Now()})
		log.Debug("stored new transaction: %s", TransactionTable.String())

		// TODO: make upstream NS configurable
		w, err := NewResponseWriter(conn, "1.1.1.1:53")
		if err != nil {
			errChan <- fmt.Errorf("failed to create a response writer: %w", err)
			return
		}

		// we skip serialization to wire for now and
		// instead reuse the original datagram
		w.Write(data)

	} else {
		// handle dns response
		log.Info("Processing DNS response")

		clientAddr, ok := TransactionTable.Load(int(m.Header.ID))
		log.Debug("loaded client address %s for transaction ID %d", clientAddr, m.Header.ID)
		if !ok {
			errChan <- fmt.Errorf("failed to find ID %d in transactions table to forward response to", m.Header.ID)
			return
		}
		w, err := NewResponseWriter(conn, clientAddr.Addr)
		if err != nil {
			errChan <- fmt.Errorf("failed to create a response writer: %w", err)
			return
		}
		w.Write(data)
		TransactionTable.Delete(int(m.Header.ID))
		log.Debug("deleted entry for transaction ID %d", m.Header.ID)
		log.Debug(TransactionTable.String())
	}

	// Here you would parse the DNS message and respond accordingly
	log.Info("Finished processing DNS data from %s", addr.String())
}
