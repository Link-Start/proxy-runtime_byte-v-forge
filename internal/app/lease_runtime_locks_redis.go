package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/redis/go-redis/v9"

	"github.com/byte-v-forge/proxy-runtime/internal/app/redisclient"
)

type redisLeaseRuntimeLocks struct {
	client *redis.Client
}

func newRedisLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (*redisLeaseRuntimeLocks, error) {
	client, err := redisclient.New(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	return &redisLeaseRuntimeLocks{client: client}, nil
}

func (s *redisLeaseRuntimeLocks) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}
