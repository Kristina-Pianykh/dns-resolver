package server

import (
	"sync"
	"time"
)

type Addr struct {
	Addr      string
	Timestamp time.Time
}

type ConnTable struct {
	mu sync.RWMutex
	m  map[int]Addr
	// timeout time.Duration
}

func (ct *ConnTable) Load(key int) (Addr, bool) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	val, ok := ct.m[key]
	return val, ok
}

func (ct *ConnTable) Store(key int, value Addr) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.m[key] = value
}

func (ct *ConnTable) Delete(key int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.m, key)
}

func InitializeConnTable() *ConnTable {
	return &ConnTable{mu: sync.RWMutex{}, m: make(map[int]Addr)}
}

// TODO: add receiver for removing expired entries
