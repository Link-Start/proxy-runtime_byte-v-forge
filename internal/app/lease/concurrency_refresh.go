package lease

import (
	"context"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type RefreshConcurrencySlotInput struct {
	Store      OrchestrationStore
	Limiter    ProviderAccountConcurrencyLimiter
	Lease      *proxyruntimev1.ProxyDynamicLease
	Limit      uint32
	DefaultTTL time.Duration
	TTLBuffer  time.Duration
}

type RefreshConcurrencySlotLimitFunc func(context.Context, *proxyruntimev1.ProxyDynamicLease, *proxyruntimev1.ProxySessionPolicy) (uint32, error)

type RefreshConcurrencySlotRunner struct {
	Store      OrchestrationStore
	Limiter    ProviderAccountConcurrencyLimiter
	DefaultTTL time.Duration
	TTLBuffer  time.Duration
	Limit      RefreshConcurrencySlotLimitFunc
}

func (r RefreshConcurrencySlotRunner) Refresh(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if !NeedsConcurrencySlotRefresh(lease, r.Limiter) {
		return nil
	}
	policy := ConcurrencyPolicy(lease)
	limit, err := r.limit(ctx, lease, policy)
	if err != nil {
		return err
	}
	return RefreshConcurrencySlot(ctx, RefreshConcurrencySlotInput{
		Store:      r.Store,
		Limiter:    r.Limiter,
		Lease:      lease,
		Limit:      limit,
		DefaultTTL: r.DefaultTTL,
		TTLBuffer:  r.TTLBuffer,
	})
}

func (r RefreshConcurrencySlotRunner) limit(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, policy *proxyruntimev1.ProxySessionPolicy) (uint32, error) {
	if r.Limit == nil {
		return 0, nil
	}
	return r.Limit(ctx, lease, policy)
}

func RefreshConcurrencySlot(ctx context.Context, input RefreshConcurrencySlotInput) error {
	if !NeedsConcurrencySlotRefresh(input.Lease, input.Limiter) {
		return nil
	}
	providerAccountID := strings.TrimSpace(input.Lease.GetProviderAccountId())
	holder := strings.TrimSpace(ConcurrencyHolder(input.Lease))
	account, err := input.Store.ProviderAccount(ctx, providerAccountID)
	if err != nil {
		return err
	}
	policy := ConcurrencyPolicy(input.Lease)
	_, err = AcquireProviderAccountConcurrencySlot(
		ctx,
		input.Limiter,
		account.GetAccountId(),
		input.Limit,
		policy,
		holder,
		ConcurrencySlotTTL(policy, input.DefaultTTL, input.TTLBuffer),
	)
	return err
}

func NeedsConcurrencySlotRefresh(lease *proxyruntimev1.ProxyDynamicLease, limiter ProviderAccountConcurrencyLimiter) bool {
	return lease != nil &&
		limiter != nil &&
		strings.TrimSpace(lease.GetProviderAccountId()) != "" &&
		strings.TrimSpace(ConcurrencyHolder(lease)) != ""
}
