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
