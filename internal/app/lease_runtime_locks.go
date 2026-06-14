package app

import (
	"context"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

const (
	leaseRuntimeLockPrefix = "proxy-runtime:lease-locks"
	leaseRuntimeLockTTL    = 2 * time.Minute
	leaseRuntimeLockRetry  = 100 * time.Millisecond
	leaseRuntimeUnlockWait = 5 * time.Second
)

type leaseRuntimeLockFunc func(context.Context) error

type leaseRuntimeLocks interface {
	Close() error
	WithAccountLock(ctx context.Context, accountID string, fn leaseRuntimeLockFunc) error
	WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseRuntimeLockFunc) error
	WithSessionListenerAllocationLock(ctx context.Context, fn leaseRuntimeLockFunc) error
}

func NewLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (leaseRuntimeLocks, error) {
	if strings.TrimSpace(cfg.RedisURL) == "" {
		return newLocalLeaseRuntimeLocks(), nil
	}
	return newRedisLeaseRuntimeLocks(ctx, cfg)
}
