package app

import (
	"context"
	"time"
)

func (l *redisLeaseRuntimeLock) renew(ctx context.Context) {
	defer close(l.done)
	defer l.scopeCancel()
	ticker := time.NewTicker(leaseRuntimeLockTTL / 3)
	defer ticker.Stop()
	if active, err := l.extend(ctx); err != nil || !active {
		if ctx.Err() == nil {
			l.lost.Store(true)
		}
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			active, err := l.extend(ctx)
			if err != nil || !active {
				if ctx.Err() == nil {
					l.lost.Store(true)
				}
				return
			}
		}
	}
}

func (l *redisLeaseRuntimeLock) extend(ctx context.Context) (bool, error) {
	return l.mutex.ExtendContext(ctx)
}
