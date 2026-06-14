package lease

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ProviderAccountConcurrencyLimiter interface {
	Acquire(context.Context, string, *proxyruntimev1.ProxySessionPolicy, uint32, string, time.Duration) (ProviderAccountConcurrencySlot, error)
	Available(context.Context, string, *proxyruntimev1.ProxySessionPolicy, uint32, string) (bool, error)
	Release(context.Context, string, *proxyruntimev1.ProxySessionPolicy, string) error
}

func AcquireProviderAccountConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, accountID string, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (ProviderAccountConcurrencySlot, error) {
	accountID = strings.TrimSpace(accountID)
	if limiter == nil {
		return NoopProviderAccountConcurrencySlot{}, nil
	}
	slot, err := limiter.Acquire(ctx, accountID, policy, limit, holder, ttl)
	if err != nil {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached: %w", accountID, ConcurrencyModeText(policy), err)
	}
	return slot, nil
}

type ProviderAccountConcurrencySlot interface {
	Release(context.Context) error
}

type NoopProviderAccountConcurrencySlot struct{}

func (NoopProviderAccountConcurrencySlot) Release(context.Context) error {
	return nil
}

func ReleaseProviderAccountConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, accountID string, policy *proxyruntimev1.ProxySessionPolicy, holder string) error {
	accountID = strings.TrimSpace(accountID)
	holder = strings.TrimSpace(holder)
	if accountID == "" || holder == "" || limiter == nil {
		return nil
	}
	return limiter.Release(ctx, accountID, policy, holder)
}

func ReleaseLeaseConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	return ReleaseProviderAccountConcurrencySlot(ctx, limiter, lease.GetProviderAccountId(), ConcurrencyPolicy(lease), ConcurrencyHolder(lease))
}

func ReleaseConcurrencySlotUnlessKept(ctx context.Context, slot ProviderAccountConcurrencySlot, keep bool) error {
	if keep || slot == nil {
		return nil
	}
	return slot.Release(ctx)
}
