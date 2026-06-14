package app

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

type redisLeaseRuntimeLock struct {
	client      *redis.Client
	key         string
	token       string
	scopeCtx    context.Context
	scopeCancel context.CancelFunc
	renewCancel context.CancelFunc
	done        chan struct{}
	lost        atomic.Bool
}

func (l *redisLeaseRuntimeLock) Unlock(ctx context.Context) error {
	if l == nil || l.client == nil || l.key == "" || l.token == "" {
		return nil
	}
	if l.renewCancel != nil {
		l.renewCancel()
	}
	if l.done != nil {
		<-l.done
	}
	return redisLeaseRuntimeUnlockScript.Run(ctx, l.client, []string{l.key}, l.token).Err()
}

func (l *redisLeaseRuntimeLock) run(fn leaseRuntimeLockFunc) error {
	if fn == nil {
		return errors.New("lease runtime lock function is required")
	}
	err := fn(l.scopeCtx)
	unlockCtx, cancel := context.WithTimeout(context.Background(), leaseRuntimeUnlockWait)
	defer cancel()
	unlockErr := l.Unlock(unlockCtx)
	if err != nil {
		return err
	}
	if l.lost.Load() {
		return errors.New("redis lease runtime lock was lost")
	}
	return unlockErr
}
