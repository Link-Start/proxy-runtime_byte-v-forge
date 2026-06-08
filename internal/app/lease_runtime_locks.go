package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	leaseRuntimeLockPrefix = "proxy-runtime:lease-locks"
	leaseRuntimeLockTTL    = 2 * time.Minute
)

type leaseRuntimeLocks interface {
	Close() error
	LockAccount(ctx context.Context, accountID string) (leaseRuntimeLock, error)
	LockProviderAccount(ctx context.Context, providerAccountID string) (leaseRuntimeLock, error)
	LockSessionListenerAllocation(ctx context.Context) (leaseRuntimeLock, error)
}

type leaseRuntimeLock interface {
	Unlock(context.Context) error
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
	client *redis.Client
	key    string
	token  string
}

func (l *redisLeaseRuntimeLock) Unlock(ctx context.Context) error {
	if l == nil || l.client == nil || l.key == "" || l.token == "" {
		return nil
	}
	return redisLeaseRuntimeUnlockScript.Run(ctx, l.client, []string{l.key}, l.token).Err()
}

func (s *redisLeaseRuntimeLocks) lock(ctx context.Context, key string) (leaseRuntimeLock, error) {
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
			return &redisLeaseRuntimeLock{client: s.client, key: redisKeyValue, token: token}, nil
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
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
