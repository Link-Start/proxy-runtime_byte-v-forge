package app

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/app/redisclient"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

const (
	leaseRuntimeLockPrefix = "proxy-runtime:lease-locks"
	leaseRuntimeLockTTL    = 2 * time.Minute
	leaseRuntimeLockRetry  = 100 * time.Millisecond
	leaseRuntimeUnlockWait = 5 * time.Second
	// 获取锁的等待由调用方 ctx 主导;leaseRuntimeLockTries 为 ctx 无截止时的兜底上限,约覆盖一个锁 TTL 的等待。
	leaseRuntimeLockTries = int(leaseRuntimeLockTTL / leaseRuntimeLockRetry)
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

func (s *redisLeaseRuntimeLocks) WithAccountLock(ctx context.Context, accountID string, fn leaseapp.LockFunc) error {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return errors.New("lease account_id is required")
	}
	return s.withLock(ctx, "account:"+accountID, fn)
}

func (s *redisLeaseRuntimeLocks) WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return errors.New("provider account id is required")
	}
	return s.withLock(ctx, "provider-account:"+providerAccountID, fn)
}

func (s *redisLeaseRuntimeLocks) WithSessionListenerAllocationLock(ctx context.Context, fn leaseapp.LockFunc) error {
	return s.withLock(ctx, "session-listener-allocation", fn)
}

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
