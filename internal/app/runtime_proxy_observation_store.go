package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/common-lib/redisx"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	proxyNodeObservationPrefix = "byte-v-forge:proxy-runtime:node-observations"
	proxyNodeObservationTTL    = 5 * time.Minute
)

type proxyNodeObservationStore struct {
	closer interface{ Close() error }
	store  *redisx.StringStore
	ttl    time.Duration
}

func newProxyNodeObservationStore(ctx context.Context, cfg config.Config) (*proxyNodeObservationStore, error) {
	client, err := redisx.NewRequiredClient(ctx, cfg.RedisURL, "PLATFORM_REDIS_URL is required")
	if err != nil {
		return nil, err
	}
	ttl := proxyNodeObservationTTL
	return &proxyNodeObservationStore{closer: client, store: redisx.NewStringStore(client, proxyNodeObservationPrefix, ttl), ttl: ttl}, nil
}

func (s *proxyNodeObservationStore) Close() error {
	if s == nil || s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

func (s *proxyNodeObservationStore) Load(ctx context.Context, key string) (*proxyruntimev1.ProxyNodeObservation, bool, error) {
	key = strings.TrimSpace(key)
	if s == nil || s.store == nil || key == "" {
		return nil, false, nil
	}
	raw, ok, err := s.store.Load(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}
	observation := &proxyruntimev1.ProxyNodeObservation{}
	if err := protojsonx.Unmarshal([]byte(raw), observation); err != nil {
		return nil, false, fmt.Errorf("parse cached proxy node observation: %w", err)
	}
	if expiresAt := observation.GetExpiresAt(); expiresAt != nil && time.Now().UTC().After(expiresAt.AsTime()) {
		return nil, false, nil
	}
	return observation, true, nil
}

func (s *proxyNodeObservationStore) Save(ctx context.Context, key string, observation *proxyruntimev1.ProxyNodeObservation) error {
	key = strings.TrimSpace(key)
	if s == nil || s.store == nil || key == "" || observation == nil {
		return nil
	}
	now := time.Now().UTC()
	if observation.GetObservedAt() == nil {
		observation.ObservedAt = timestamppb.New(now)
	}
	if observation.GetExpiresAt() == nil {
		observation.ExpiresAt = timestamppb.New(now.Add(s.ttl))
	}
	data, err := protojsonx.Marshal(observation)
	if err != nil {
		return err
	}
	return s.store.SaveTTL(ctx, key, string(data), s.ttl)
}
