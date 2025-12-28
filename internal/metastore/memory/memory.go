package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Metastore struct {
	mu        sync.RWMutex
	store     map[string][]byte
	keyPrefix string
}

func New(keyPrefix string) *Metastore {
	return &Metastore{
		store:     make(map[string][]byte),
		keyPrefix: keyPrefix,
	}
}

func (m *Metastore) Get(ctx context.Context, uniqueID string, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.store[m.dataKey(uniqueID, key)]
	if !ok {
		return nil, errors.New("data not found")
	}

	return data, nil
}

func (m *Metastore) Set(ctx context.Context, uniqueID string, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dataBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to json marshal value: %w", err)
	}

	m.store[m.dataKey(uniqueID, key)] = dataBytes
	return nil
}

func (m *Metastore) SetWithTTL(ctx context.Context, uniqueID string, key string, value any, ttl time.Duration) error {
	// TODO: Implement TTL
	slog.InfoContext(ctx, "Memory metastore does not support TTL")
	return m.Set(ctx, uniqueID, key, value)
}

func (m *Metastore) Delete(ctx context.Context, uniqueID string, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, m.dataKey(uniqueID, key))
	return nil
}

func (m *Metastore) Exists(ctx context.Context, uniqueID string, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.store[m.dataKey(uniqueID, key)]
	return ok, nil
}

func (m *Metastore) dataKey(uniqueID string, key string) string {
	return fmt.Sprintf("%s::%s::%s", m.keyPrefix, uniqueID, key)
}
