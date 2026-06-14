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

func ReleaseConcurrencySlotUnlessKept(ctx context.Context, slot ProviderAccountConcurrencySlot, keep bool) error {
	if keep || slot == nil {
		return nil
	}
	return slot.Release(ctx)
}
