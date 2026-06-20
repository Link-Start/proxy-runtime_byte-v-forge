package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/app/redisclient"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type redisLeaseRuntimeLocks struct {
	client *redis.Client
	rs     *redsync.Redsync
}

func newRedisLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (*redisLeaseRuntimeLocks, error) {
	client, err := redisclient.New(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	return &redisLeaseRuntimeLocks{
		client: client,
		rs:     redsync.New(goredis.NewPool(client)),
	}, nil
}

func (s *redisLeaseRuntimeLocks) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}
