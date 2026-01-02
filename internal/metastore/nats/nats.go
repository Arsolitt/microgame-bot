package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"microgame-bot/internal/metastore"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultMaxBytes = 1024 * 1024 * 10 // 10MB
	defaultTTL      = 30 * 24 * time.Hour
)

type Metastore struct {
	bucket jetstream.KeyValue
}

func New(ctx context.Context, conn *nats.Conn, bucketName string, replicas int) (*Metastore, error) {
	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}
	bucket, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:         bucketName,
		MaxBytes:       defaultMaxBytes,
		TTL:            defaultTTL,
		Storage:        jetstream.FileStorage,
		Replicas:       replicas,
		Compression:    true,
		LimitMarkerTTL: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create or update keyvalue bucket: %w", err)
	}
	return &Metastore{bucket: bucket}, nil
}

func (m *Metastore) Get(ctx context.Context, uniqueID string, key string) ([]byte, error) {
	data, err := m.bucket.Get(ctx, m.dataKey(uniqueID, key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, metastore.ErrKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}

	return data.Value(), nil
}

func (m *Metastore) GetString(ctx context.Context, uniqueID string, key string) (string, error) {
	data, err := m.bucket.Get(ctx, m.dataKey(uniqueID, key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return "", metastore.ErrKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get data: %w", err)
	}
	return string(data.Value()), nil
}

func (m *Metastore) Set(ctx context.Context, uniqueID string, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to json marshal value: %w", err)
	}

	_, err = m.bucket.Put(ctx, m.dataKey(uniqueID, key), data)
	if err != nil {
		return fmt.Errorf("failed to put data: %w", err)
	}
	return nil
}

func (m *Metastore) SetString(ctx context.Context, uniqueID string, key string, value string) error {
	_, err := m.bucket.Put(ctx, m.dataKey(uniqueID, key), []byte(value))
	if err != nil {
		return fmt.Errorf("failed to put data: %w", err)
	}
	return nil
}

func (m *Metastore) SetWithTTL(ctx context.Context, uniqueID string, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to json marshal value: %w", err)
	}
	return m.сreateOrUpdate(ctx, m.dataKey(uniqueID, key), data, ttl)
}

func (m *Metastore) SetStringWithTTL(
	ctx context.Context,
	uniqueID string,
	key string,
	value string,
	ttl time.Duration,
) error {
	return m.сreateOrUpdate(ctx, m.dataKey(uniqueID, key), []byte(value), ttl)
}

func (m *Metastore) Delete(ctx context.Context, uniqueID string, key string) error {
	err := m.bucket.Delete(ctx, m.dataKey(uniqueID, key))
	if err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}
	return nil
}

func (m *Metastore) Exists(ctx context.Context, uniqueID string, key string) (bool, error) {
	_, err := m.bucket.Get(ctx, m.dataKey(uniqueID, key))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get data: %w", err)
	}
	return true, nil
}

func (m *Metastore) dataKey(uniqueID string, key string) string {
	return fmt.Sprintf("%s__%s", uniqueID, key)
}

func (m *Metastore) сreateOrUpdate(ctx context.Context, dataKey string, data []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = defaultTTL
	}

	keyEntry, err := m.bucket.Get(ctx, dataKey)
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		return fmt.Errorf("failed to get data: %w", err)
	} else if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		_, err = m.bucket.Create(ctx, dataKey, data, jetstream.KeyTTL(ttl))
		if err != nil {
			return fmt.Errorf("failed to create data: %w", err)
		}
	} else {
		_, err = m.bucket.Update(ctx, dataKey, data, keyEntry.Revision())
		if err != nil {
			return fmt.Errorf("failed to update data: %w", err)
		}
	}
	return nil
}
