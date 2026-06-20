package app

import (
	"context"
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/app/redisclient"
	"github.com/go-redsync/redsync/v4"
)

func (s *redisLeaseRuntimeLocks) withLock(ctx context.Context, key string, fn leaseapp.LockFunc) error {
	lock, err := s.lock(ctx, key)
	if err != nil {
		return err
	}
	return lock.run(fn)
}

func (s *redisLeaseRuntimeLocks) lock(ctx context.Context, key string) (*redisLeaseRuntimeLock, error) {
	if s == nil || s.rs == nil {
		return nil, errors.New("redis lease lock client is not configured")
	}
	redisKeyValue, ok := redisclient.Key(leaseRuntimeLockPrefix, key)
	if !ok {
		return nil, errors.New("redis lease lock key is required")
	}
	mutex := s.rs.NewMutex(redisKeyValue,
		redsync.WithExpiry(leaseRuntimeLockTTL),
		redsync.WithTries(leaseRuntimeLockTries),
		redsync.WithRetryDelay(leaseRuntimeLockRetry),
	)
	if err := mutex.LockContext(ctx); err != nil {
		return nil, err
	}
	scopeCtx, scopeCancel := context.WithCancel(ctx)
	renewCtx, renewCancel := context.WithCancel(context.Background())
	lock := &redisLeaseRuntimeLock{
		mutex:       mutex,
		scopeCtx:    scopeCtx,
		scopeCancel: scopeCancel,
		renewCancel: renewCancel,
		done:        make(chan struct{}),
	}
	go lock.renew(renewCtx)
	return lock, nil
}
