package kv

import (
	"context"
	"fmt"
	"microgame-bot/internal/core"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func GetNatsConn(ctx context.Context, cfg core.NatsConfig) (*nats.Conn, error) {
	sig, err := nkeys.FromSeed([]byte(cfg.NKeySeed))
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}
	return nats.Connect(cfg.URL, nats.Nkey(cfg.NKeyPublic, nats.SignatureHandler(func(nonce []byte) ([]byte, error) {
		sigData, err := sig.Sign(nonce)
		sig.Wipe()
		return sigData, err
	})))
}
