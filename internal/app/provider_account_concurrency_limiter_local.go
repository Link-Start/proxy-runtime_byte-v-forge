package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type localProviderAccountConcurrencyLimiter struct {
	mu    sync.Mutex
	slots map[string]map[string]time.Time
}

type localProviderAccountConcurrencySlot struct {
	limiter   *localProviderAccountConcurrencyLimiter
	accountID string
	policy    *proxyruntimev1.ProxySessionPolicy
	holder    string
}

func newLocalProviderAccountConcurrencyLimiter() *localProviderAccountConcurrencyLimiter {
	return &localProviderAccountConcurrencyLimiter{slots: map[string]map[string]time.Time{}}
}

func (l *localProviderAccountConcurrencyLimiter) Close() error { return nil }

func (l *localProviderAccountConcurrencyLimiter) Available(ctx context.Context, accountID string, policy *proxyruntimev1.ProxySessionPolicy, limit uint32, holder string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	key, cleanHolder, err := localConcurrencyKey(accountID, policy, holder)
	if err != nil {
		return false, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().UTC()
	l.purgeExpiredLocked(key, now)
	if cleanHolder != "" && cleanHolder != "_probe" {
		if _, exists := l.slots[key][cleanHolder]; exists {
			return true, nil
		}
	}
	return len(l.slots[key]) < int(limit), nil
}

func (l *localProviderAccountConcurrencyLimiter) Acquire(ctx context.Context, accountID string, policy *proxyruntimev1.ProxySessionPolicy, limit uint32, holder string, ttl time.Duration) (providerAccountConcurrencySlot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key, cleanHolder, err := localConcurrencyKey(accountID, policy, holder)
	if err != nil {
		return nil, err
	}
	ttl = effectiveProviderAccountConcurrencySlotTTL(ttl)
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().UTC()
	l.purgeExpiredLocked(key, now)
	if _, exists := l.slots[key][cleanHolder]; !exists && len(l.slots[key]) >= int(limit) {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached", strings.TrimSpace(accountID), providerAccountConcurrencyModeText(policy))
	}
	l.slots[key][cleanHolder] = now.Add(ttl)
	return &localProviderAccountConcurrencySlot{limiter: l, accountID: strings.TrimSpace(accountID), policy: cloneConcurrencyPolicy(policy), holder: cleanHolder}, nil
}

func (l *localProviderAccountConcurrencyLimiter) Release(ctx context.Context, accountID string, policy *proxyruntimev1.ProxySessionPolicy, holder string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key, cleanHolder, err := localConcurrencyKey(accountID, policy, holder)
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.slots[key], cleanHolder)
	if len(l.slots[key]) == 0 {
		delete(l.slots, key)
	}
	return nil
}

func (s *localProviderAccountConcurrencySlot) Release(ctx context.Context) error {
	if s == nil || s.limiter == nil {
		return nil
	}
	return s.limiter.Release(ctx, s.accountID, s.policy, s.holder)
}

func (l *localProviderAccountConcurrencyLimiter) purgeExpiredLocked(key string, now time.Time) {
	items := l.slots[key]
	if items == nil {
		items = map[string]time.Time{}
		l.slots[key] = items
		return
	}
	for holder, expiresAt := range items {
		if !expiresAt.After(now) {
			delete(items, holder)
		}
	}
}

func localConcurrencyKey(accountID string, policy *proxyruntimev1.ProxySessionPolicy, holder string) (string, string, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return "", "", fmt.Errorf("provider account id is required")
	}
	holder = strings.TrimSpace(holder)
	if holder == "" {
		holder = "_probe"
	}
	return accountID + ":" + providerAccountConcurrencyModeText(policy), holder, nil
}
