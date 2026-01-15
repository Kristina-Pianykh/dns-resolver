package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
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
	UDPServer *UDPServer
	// TCPServer *TCPServer
}

type Message struct {
	Data []byte
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

	udpSrv, err := NewUDPServer(&srvCfg.UDPCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP server: %w", err)
	}

	srv := Server{
		Cfg:       &srvCfg,
		UDPServer: udpSrv,
	}
	return &srv, nil
}

func (s *Server) Start(ctx context.Context) error {
	log.Info("starting server...")
	if s.UDPServer == nil {
		return errors.New("failed to start server: UDP listener is not initialized")
	}

	go func() {
		if err := s.UDPServer.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to start UDP server: %v\n", err)
			os.Exit(1)
		}
	}()
	return nil
}

type UDPServer struct {
	Config      *UDPConfig
	Conn        *net.UDPConn
	UDPSessions *SafeMap
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
		Config:      cfg,
		Conn:        UDPConn,
		UDPSessions: &SafeMap{mu: sync.RWMutex{}, m: make(map[string]time.Time)},
	}

	return &srv, nil
}

func (s *UDPServer) Start(ctx context.Context) error {
	log.Info("starting UDP server...")
	if s.Conn == nil {
		return errors.New("UDP connection is not initialized")
	}

	// if parent context is done, close the UDP connection
	go func() {
		<-ctx.Done()
		_ = s.Conn.Close()
	}()

	buf := make([]byte, s.Config.MaxBufferSize)
	for {
		n, addr, err := s.Conn.ReadFromUDP(buf[0:])
		if err != nil {
			select {
			case <-ctx.Done():
				log.Info("UDP server shutting down")
				return nil
			default:
				// might send to error channel instead
				log.Error("error reading from UDP:", err)
				continue
			}
		}
		log.Info("Received a packet from", addr.String())

		// copy data to avoid overwriting in next read
		data := make([]byte, s.Config.MaxBufferSize)
		copy(data, buf[:n])

		// TODO: set a timeout for the session
		go func(data []byte, addr *net.UDPAddr) {
			// Per-packet timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, s.Config.Timeout)
			defer cancel()

			key := addr.String()
			s.UDPSessions.Store(key, time.Now())
			defer s.UDPSessions.Delete(key)

			DNSProcess(data, addr, s.Conn, timeoutCtx)
		}(data, addr)

		// Reset buffer for next read
		buf = buf[:0]
	}
}

func DNSProcess(data []byte, addr *net.UDPAddr, conn *net.UDPConn, ctx context.Context) {
	// Placeholder for DNS processing logic
	log.Info("Processing DNS data from", addr.String())
	p, err := parser.NewParser(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = p.ParseMessage()
	if err != nil {
		log.Error("failed to parse DNS message:", err)
		return
	}

	if p.Message.IsQuery() {
		// resolve query
		// make a dns query
		// block, wait for response or timeout
	} else {
		// handle response
		// write dns in binary as udpo back to addr
		_, err := conn.WriteToUDP(data, addr)
		if err != nil {
			log.Error("failed to write DNS response to", addr.String(), ":", err)
			return
		}
	}

	// Here you would parse the DNS message and respond accordingly
	log.Info("Finished processing DNS data from", addr.String())
}

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]time.Time
}

func (sm *SafeMap) Load(key string) (time.Time, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	val, ok := sm.m[key]
	return val, ok
}

func (sm *SafeMap) Store(key string, value time.Time) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

func (sm *SafeMap) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.m, key)
}
