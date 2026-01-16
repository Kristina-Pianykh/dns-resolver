package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var TestUDPCfg = UDPConfig{
	Addr:          "127.0.0.1",
	Port:          0, // random available port
	MaxBufferSize: 512,
}

func TestUDPServerCreation(t *testing.T) {
	udpSrv, err := NewUDPServer(&TestUDPCfg)
	assert.NoError(t, err)
	assert.NotNil(t, udpSrv)
	assert.NotNil(t, udpSrv.Conn)
	assert.Equal(t, "udp", udpSrv.GetNet())
}

func TestUDPServerLifecycleViaShutdown(t *testing.T) {
	udpSrv, err := NewUDPServer(&TestUDPCfg)
	assert.NoError(t, err)
	assert.NotNil(t, udpSrv)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var wg sync.WaitGroup
	wg.Go(func() {
		err := udpSrv.Start(ctx)
		assert.NoError(t, err)
	})

	time.Sleep(time.Second * 2)

	err = udpSrv.Shutdown(context.Background())
	assert.NoError(t, err)

	wg.Wait()

	// Shutdown should not happen before the timeout
	// to test that shutting down doesn't hang
	assert.Nil(t, ctx.Err())
}

func TestUDPServerLifecycleViaContext(t *testing.T) {
	udpSrv, err := NewUDPServer(&TestUDPCfg)
	assert.NoError(t, err)
	assert.NotNil(t, udpSrv)

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	wg.Go(func() {
		err = udpSrv.Start(ctx)
		assert.NoError(t, err)
	})

	wg.Wait()

	assert.NotNil(t, ctx.Err())
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}

func TestUDPMessage(t *testing.T) {
	udpSrv, err := NewUDPServer(&TestUDPCfg)
	assert.NoError(t, err)
	assert.NotNil(t, udpSrv)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Go(func() {
		err = udpSrv.Start(ctx)
		assert.NoError(t, err)
	})

	time.Sleep(time.Second * 1)

	clients := make([]*UDPClient, 0)
	for i := 0; i < 5; i++ {
		client, err := NewUDPClient(udpSrv.Conn.LocalAddr().String())
		assert.NoError(t, err)
		assert.NotNil(t, client)

		clients = append(clients, client)
	}

	hexQuery := "7e4e01000001000000000000076e69636b6c6173077365646c6f636b0378797a0000010001"
	query, err := hex.DecodeString(hexQuery)
	assert.NoError(t, err)

	for _, client := range clients {
		resp, err := client.SendAndReceive(query, 512)
		assert.NoError(t, err)

		assert.NotEmpty(t, resp)
		fmt.Println("Response: ", string(resp))
	}

	cancel()

	wg.Wait()
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
