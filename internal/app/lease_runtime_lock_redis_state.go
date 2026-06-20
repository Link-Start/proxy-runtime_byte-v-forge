package app

import (
	"context"
	"errors"
	"sync/atomic"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/go-redsync/redsync/v4"
)

type redisLeaseRuntimeLock struct {
	mutex       *redsync.Mutex
	scopeCtx    context.Context
	scopeCancel context.CancelFunc
	renewCancel context.CancelFunc
	done        chan struct{}
	lost        atomic.Bool
}

func (l *redisLeaseRuntimeLock) Unlock(ctx context.Context) error {
	if l == nil || l.mutex == nil {
		return nil
	}
	if l.renewCancel != nil {
		l.renewCancel()
	}
	if l.done != nil {
		<-l.done
	}
	if _, err := l.mutex.UnlockContext(ctx); err != nil {
		return err
	}
	return nil
}

func (l *redisLeaseRuntimeLock) run(fn leaseapp.LockFunc) error {
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
