package server

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMapWriteAccess(t *testing.T) {
	ct := InitializeConnTable()
	N := 100

	var wg sync.WaitGroup
	for i := range N {
		wg.Go(func() {
			ct.Store(i, Addr{
				Addr:      fmt.Sprintf("addr_%d", i),
				Timestamp: time.Now(),
			})
		})
	}
	wg.Wait()

	assert.Equal(t, N, len(ct.m))
	for i := range N {
		val, ok := ct.Load(i)
		assert.True(t, ok)
		assert.Equal(t, fmt.Sprintf("addr_%d", i), val.Addr)
	}
}
