package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"server/pkg/log"

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
	InitTransactionsTable()

	assert.NoError(t, err)
	assert.NotNil(t, udpSrv)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		err = udpSrv.Start(ctx, errCh)
		assert.NoError(t, err)
	})

	fmt.Println("launched server")
	time.Sleep(time.Second * 1)

	clients := make([]*UDPClient, 0)
	for i := range 5 {
		client, err := NewUDPClient(udpSrv.Conn.LocalAddr().String())
		fmt.Printf("launching client %d\n", i)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		clients = append(clients, client)
	}

	hexQuery := "45dc010000010000000000000377777707796f757475626503636f6d0000010001"
	query, err := hex.DecodeString(hexQuery)
	assert.NoError(t, err)

	fmt.Println("we are here")
	for _, client := range clients {
		wg.Go(func() {
			resp, err := client.SendAndReceive(query, 512)
			defer client.Close()

			fmt.Printf("sent and received\n")
			assert.NoError(t, err)

			assert.NotEmpty(t, resp)
			log.Debug("Response: ", string(resp))
		})
	}

	wg.Wait()
	cancel()
	fmt.Println("we are here")
	assert.Empty(t, errCh)
}
