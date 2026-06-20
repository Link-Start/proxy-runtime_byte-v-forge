package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/app/redisclient"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	providerAccountConcurrencyKeyPrefix      = "proxy-runtime:provider-account-concurrency"
	defaultProviderAccountConcurrencySlotTTL = 15 * time.Minute
)

type redisProviderAccountConcurrencyLimiter struct {
	client *redis.Client
	clock  clock.Clock
}

type providerAccountConcurrencyRuntime interface {
	leaseapp.ProviderAccountConcurrencyLimiter
	Close() error
}

type redisProviderAccountConcurrencySlot struct {
	limiter   *redisProviderAccountConcurrencyLimiter
	accountID string
	policy    *proxyruntimev1.ProxySessionPolicy
	holder    string
}

func NewProviderAccountConcurrencyLimiter(ctx context.Context, cfg config.Config, clk clock.Clock) (providerAccountConcurrencyRuntime, error) {
	client, err := redisclient.New(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	return &redisProviderAccountConcurrencyLimiter{client: client, clock: clk}, nil
}

func (l *redisProviderAccountConcurrencyLimiter) Close() error {
	if l == nil || l.client == nil {
		return nil
	}
	return l.client.Close()
}

func (l *redisProviderAccountConcurrencyLimiter) Available(ctx context.Context, accountID string, policy *proxyruntimev1.ProxySessionPolicy, limit uint32, holder string) (bool, error) {
	redisKey, cleanHolder, err := l.key(accountID, policy, holder)
	if err != nil {
		return false, err
	}
	result, err := providerAccountConcurrencyAvailableScript.Run(ctx, l.client, []string{redisKey}, cleanHolder, strconv.FormatUint(uint64(limit), 10), strconv.FormatInt(l.clock.Now().UnixMilli(), 10)).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (l *redisProviderAccountConcurrencyLimiter) Acquire(ctx context.Context, accountID string, policy *proxyruntimev1.ProxySessionPolicy, limit uint32, holder string, ttl time.Duration) (leaseapp.ProviderAccountConcurrencySlot, error) {
	redisKey, cleanHolder, err := l.key(accountID, policy, holder)
	if err != nil {
		return nil, err
	}
	ttl = effectiveProviderAccountConcurrencySlotTTL(ttl)
	result, err := providerAccountConcurrencyAcquireScript.Run(ctx, l.client, []string{redisKey}, cleanHolder, strconv.FormatUint(uint64(limit), 10), strconv.FormatInt(ttl.Milliseconds(), 10), strconv.FormatInt(l.clock.Now().UnixMilli(), 10)).Int()
	if err != nil {
		return nil, err
	}
	if result != 1 {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached", strings.TrimSpace(accountID), leaseapp.ConcurrencyModeText(policy))
	}
	return &redisProviderAccountConcurrencySlot{limiter: l, accountID: strings.TrimSpace(accountID), policy: cloneConcurrencyPolicy(policy), holder: cleanHolder}, nil
}

func (l *redisProviderAccountConcurrencyLimiter) Release(ctx context.Context, accountID string, policy *proxyruntimev1.ProxySessionPolicy, holder string) error {
	redisKey, cleanHolder, err := l.key(accountID, policy, holder)
	if err != nil {
		return err
	}
	return providerAccountConcurrencyReleaseScript.Run(ctx, l.client, []string{redisKey}, cleanHolder).Err()
}

func (s *redisProviderAccountConcurrencySlot) Release(ctx context.Context) error {
	if s == nil || s.limiter == nil {
		return nil
	}
	return s.limiter.Release(ctx, s.accountID, s.policy, s.holder)
}

func (l *redisProviderAccountConcurrencyLimiter) key(accountID string, policy *proxyruntimev1.ProxySessionPolicy, holder string) (string, string, error) {
	if l == nil || l.client == nil {
		return "", "", errors.New("provider account concurrency cache is not configured")
	}
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return "", "", errors.New("provider account id is required")
	}
	holder = strings.TrimSpace(holder)
	if holder == "" {
		holder = "_probe"
	}
	key, ok := redisclient.Key(providerAccountConcurrencyKeyPrefix, accountID+":"+leaseapp.ConcurrencyModeText(policy))
	if !ok {
		return "", "", errors.New("provider account concurrency key is required")
	}
	return key, holder, nil
}

func effectiveProviderAccountConcurrencySlotTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return defaultProviderAccountConcurrencySlotTTL
	}
	if ttl < time.Minute {
		return time.Minute
	}
	return ttl
}

func cloneConcurrencyPolicy(policy *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	if policy == nil {
		return nil
	}
	return &proxyruntimev1.ProxySessionPolicy{
		Mode:         policy.GetMode(),
		RotationMode: policy.GetRotationMode(),
	}
}

var providerAccountConcurrencyAvailableScript = redis.NewScript(`
local key = KEYS[1]
local holder = ARGV[1]
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
redis.call("ZREMRANGEBYSCORE", key, "-inf", now)
if holder ~= "" and holder ~= "_probe" and redis.call("ZSCORE", key, holder) then
  return 1
end
if redis.call("ZCARD", key) < limit then
  return 1
end
return 0
`)

var providerAccountConcurrencyAcquireScript = redis.NewScript(`
local key = KEYS[1]
local holder = ARGV[1]
local limit = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local expires_at = now + ttl
redis.call("ZREMRANGEBYSCORE", key, "-inf", now)
if redis.call("ZSCORE", key, holder) then
  redis.call("ZADD", key, expires_at, holder)
  redis.call("PEXPIRE", key, ttl)
  return 1
end
if redis.call("ZCARD", key) >= limit then
  return 0
end
redis.call("ZADD", key, expires_at, holder)
redis.call("PEXPIRE", key, ttl)
return 1
`)

var providerAccountConcurrencyReleaseScript = redis.NewScript(`
local key = KEYS[1]
local holder = ARGV[1]
redis.call("ZREM", key, holder)
if redis.call("ZCARD", key) == 0 then
  redis.call("DEL", key)
end
return 1
`)
