package metastore

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type IMetastore interface {
	IMetastoreGetter
	IMetastoreSetter
	IMetastoreDeleter
}

type IMetastoreGetter interface {
	Get(ctx context.Context, uniqueID string, key string) ([]byte, error)
	GetString(ctx context.Context, uniqueID string, key string) (string, error)
	Exists(ctx context.Context, uniqueID string, key string) (bool, error)
}

type IMetastoreSetter interface {
	Set(ctx context.Context, uniqueID string, key string, value any) error
	SetString(ctx context.Context, uniqueID string, key string, value string) error
	SetWithTTL(ctx context.Context, uniqueID string, key string, value any, ttl time.Duration) error
	SetStringWithTTL(ctx context.Context, uniqueID string, key string, value string, ttl time.Duration) error
}

type IMetastoreDeleter interface {
	Delete(ctx context.Context, uniqueID string, key string) error
}

func TypedJSONMeta[T any](ctx context.Context, mg IMetastoreGetter, uniqueID string, key string) (T, error) {
	var zero T
	data, err := mg.Get(ctx, uniqueID, key)
	if err != nil {
		return zero, err
	}

	var typedData T
	err = json.Unmarshal(data, &typedData)
	if err != nil {
		return zero, err
	}

	return typedData, nil
}
