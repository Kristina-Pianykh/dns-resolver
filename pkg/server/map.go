package server

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var TransactionTable *TransationsTable

type Addr struct {
	Addr      string
	Timestamp time.Time
}

type TransationsTable struct {
	mu sync.RWMutex
	m  map[int]Addr
	// timeout time.Duration
}

func (tt *TransationsTable) Load(key int) (Addr, bool) {
	if tt == nil || tt.m == nil {
		log.Fatal("TransactionsTable is not initialized")
		return Addr{}, false
	}

	tt.mu.RLock()
	defer tt.mu.RUnlock()
	val, ok := tt.m[key]
	return val, ok
}

func (tt *TransationsTable) Store(key int, value Addr) {
	if tt == nil || tt.m == nil {
		log.Fatal("TransactionsTable is not initialized")
		return
	}
	tt.mu.Lock()

	defer tt.mu.Unlock()
	tt.m[key] = value
}

func (tt *TransationsTable) Delete(key int) {
	if tt == nil || tt.m == nil {
		log.Fatal("TransactionsTable is not initialized")
		return
	}

	tt.mu.Lock()
	defer tt.mu.Unlock()
	delete(tt.m, key)
}

func InitTransactionsTable() {
	TransactionTable = &TransationsTable{mu: sync.RWMutex{}, m: make(map[int]Addr)}
}

func (tt *TransationsTable) String() string {
	if tt == nil {
		log.Fatal("TransactionsTable is not initialized")
		return ""
	}
	var sb strings.Builder
	for k, v := range tt.m {
		sb.WriteString(fmt.Sprintf("ID: %d, %v", k, v))
	}
	return sb.String()
}

// TODO: add receiver for removing expired entries
