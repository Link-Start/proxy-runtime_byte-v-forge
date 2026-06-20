package app

import (
	"context"
	"errors"
	"time"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/app/redisclient"
)

func (s *redisLeaseRuntimeLocks) withLock(ctx context.Context, key string, fn leaseapp.LockFunc) error {
	lock, err := s.lock(ctx, key)
	if err != nil {
		return err
	}
	return lock.run(fn)
}

func (s *redisLeaseRuntimeLocks) lock(ctx context.Context, key string) (*redisLeaseRuntimeLock, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("redis lease lock client is not configured")
	}
	redisKeyValue, ok := redisclient.Key(leaseRuntimeLockPrefix, key)
	if !ok {
		return nil, errors.New("redis lease lock key is required")
	}
	token, err := redisLeaseRuntimeLockToken()
	if err != nil {
		return nil, err
	}
	for {
		locked, err := s.client.SetNX(ctx, redisKeyValue, token, leaseRuntimeLockTTL).Result()
		if err != nil {
			return nil, err
		}
		if locked {
			scopeCtx, scopeCancel := context.WithCancel(ctx)
			renewCtx, renewCancel := context.WithCancel(context.Background())
			lock := &redisLeaseRuntimeLock{client: s.client, key: redisKeyValue, token: token, scopeCtx: scopeCtx, scopeCancel: scopeCancel, renewCancel: renewCancel, done: make(chan struct{})}
			go lock.renew(renewCtx)
			return lock, nil
		}
		timer := time.NewTimer(leaseRuntimeLockRetry)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
