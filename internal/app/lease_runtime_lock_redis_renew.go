package app

import (
	"context"
	"fmt"
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
	result, err := redisLeaseRuntimeExtendScript.Run(ctx, l.client, []string{l.key}, l.token, leaseRuntimeLockTTL.Milliseconds()).Int()
	if err != nil {
		return false, fmt.Errorf("extend redis lease runtime lock: %w", err)
	}
	return result == 1, nil
}
