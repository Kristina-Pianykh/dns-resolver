package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	logging "server/pkg/log"
	"server/pkg/parser"
	"server/pkg/server"
)

var (
	signals = make(chan os.Signal, 1)
	done    = make(chan bool, 1)
)

const shutdownTimeout = 5 * time.Second

func parse(data []byte) error {
	p, err := parser.NewParser(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = p.ParseMessage()
	if err != nil {
		return fmt.Errorf("failed to parse message: %v", err)
	}
	fmt.Println(p.Message.Header.String())
	fmt.Println(p.Message.Question.String())

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	server.InitTransactionsTable()

	srvErrChan := make(chan error, 1)
	procErrChan := make(chan error, 10)

	// Resolve the string address to a UDP address
	srv, err := server.NewServer()
	if err != nil {
		logging.Error("failed to create server: ", err)
		os.Exit(1)
	}

	srv.Start(ctx, srvErrChan, procErrChan)

	go func() {
		select {
		case <-signals:
			logging.Info("Terminating...")

			// Cancel background operations (periodic refresh, etc.)
			cancelFn()

			// Create timeout context for graceful shutdown
			stopCtx, stopCancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer stopCancel()

			srv.Stop(stopCtx)
			done <- true

		case err := <-srvErrChan:
			logging.Error("server start failed: ", err)
			done <- true
		case err := <-procErrChan:
			logging.Error(err.Error())
		}
	}()

	<-done
}
