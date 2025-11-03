package blinkdb

import (
	"sync"
)

// KV represents a thread-safe key-value store
type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewKV creates a new key-value store instance
func NewKV() *KV {
	return &KV{
		data: make(map[string][]byte),
	}
}

// Set stores a key-value pair
func (kv *KV) Set(key, val []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[string(key)] = val
	return nil
}

// Get retrieves a value by key
func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.data[string(key)]
	return val, ok
}

// Delete removes a key-value pair
func (kv *KV) Delete(key []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.data, string(key))
	return nil
}
