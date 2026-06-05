package app

import (
	"context"
	"time"

	"github.com/byte-v-forge/common-lib/redisx"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

const (
	leaseRuntimeLockPrefix = "byte-v-forge:proxy-runtime:lease-locks"
	leaseRuntimeLockTTL    = 2 * time.Minute
)

type leaseRuntimeLocks interface {
	Close() error
	LockAccount(ctx context.Context, accountID string) (leaseRuntimeLock, error)
	LockProviderAccount(ctx context.Context, providerAccountID string) (leaseRuntimeLock, error)
	LockSessionListenerAllocation(ctx context.Context) (leaseRuntimeLock, error)
}

type leaseRuntimeLock interface {
	Unlock(context.Context) error
}

type redisLeaseRuntimeLocks struct {
	closer interface{ Close() error }
	locker *redisx.BestEffortLocker
}

func NewLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (leaseRuntimeLocks, error) {
	return newRedisLeaseRuntimeLocks(ctx, cfg)
}

func newRedisLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (*redisLeaseRuntimeLocks, error) {
	client, err := redisx.NewRequiredClient(ctx, cfg.RedisURL, "PLATFORM_REDIS_URL is required")
	if err != nil {
		return nil, err
	}
	return &redisLeaseRuntimeLocks{closer: client, locker: redisx.NewBestEffortLocker(client, leaseRuntimeLockPrefix, leaseRuntimeLockTTL, 100*time.Millisecond)}, nil
}

func (s *redisLeaseRuntimeLocks) Close() error {
	if s == nil || s.closer == nil {
		return nil
	}
	return s.closer.Close()
}
