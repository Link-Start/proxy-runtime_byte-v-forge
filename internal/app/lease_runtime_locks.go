package app

import (
	"context"
	"time"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

const (
	leaseRuntimeLockPrefix = "proxy-runtime:lease-locks"
	leaseRuntimeLockTTL    = 2 * time.Minute
	leaseRuntimeLockRetry  = 100 * time.Millisecond
	leaseRuntimeUnlockWait = 5 * time.Second
)

type leaseRuntimeLocks interface {
	Close() error
	WithAccountLock(ctx context.Context, accountID string, fn leaseapp.LockFunc) error
	WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error
	WithSessionListenerAllocationLock(ctx context.Context, fn leaseapp.LockFunc) error
}

func NewLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (leaseRuntimeLocks, error) {
	return newRedisLeaseRuntimeLocks(ctx, cfg)
}
