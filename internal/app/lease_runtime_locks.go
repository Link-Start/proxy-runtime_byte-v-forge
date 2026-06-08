package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/redis/go-redis/v9"
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

type redisLeaseRuntimeLocks struct {
	client *redis.Client
}

func NewLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (leaseRuntimeLocks, error) {
	if strings.TrimSpace(cfg.RedisURL) == "" {
		return newLocalLeaseRuntimeLocks(), nil
	}
	return newRedisLeaseRuntimeLocks(ctx, cfg)
}

func newRedisLeaseRuntimeLocks(ctx context.Context, cfg config.Config) (*redisLeaseRuntimeLocks, error) {
	client, err := newRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	return &redisLeaseRuntimeLocks{client: client}, nil
}

func (s *redisLeaseRuntimeLocks) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

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

func (s *redisLeaseRuntimeLocks) withLock(ctx context.Context, key string, fn leaseRuntimeLockFunc) error {
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
	redisKeyValue, ok := redisKey(leaseRuntimeLockPrefix, key)
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

func redisLeaseRuntimeLockToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

var redisLeaseRuntimeUnlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

var redisLeaseRuntimeExtendScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)
