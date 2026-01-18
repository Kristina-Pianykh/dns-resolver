package server

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMapWriteAccess(t *testing.T) {
	InitTransactionsTable()
	N := 100

	var wg sync.WaitGroup
	for i := range N {
		wg.Go(func() {
			TransactionTable.Store(i, Addr{
				Addr:      fmt.Sprintf("addr_%d", i),
				Timestamp: time.Now(),
			})
		})
	}
	wg.Wait()

	assert.Equal(t, N, len(TransactionTable.m))
	for i := range N {
		val, ok := TransactionTable.Load(i)
		assert.True(t, ok)
		assert.Equal(t, fmt.Sprintf("addr_%d", i), val.Addr)
	}
}
