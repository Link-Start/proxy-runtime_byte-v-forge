package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type AcquireAttemptConcurrencyInput struct {
	Limiter    ProviderAccountConcurrencyLimiter
	AccountID  string
	Limit      uint32
	Policy     *proxyruntimev1.ProxySessionPolicy
	LeaseID    string
	DefaultTTL time.Duration
	TTLBuffer  time.Duration
}

func AcquireAttemptConcurrencySlot(ctx context.Context, input AcquireAttemptConcurrencyInput) (ProviderAccountConcurrencySlot, string, error) {
	holder := HolderForLeaseID(input.LeaseID)
	slot, err := AcquireProviderAccountConcurrencySlot(
		ctx,
		input.Limiter,
		input.AccountID,
		input.Limit,
		input.Policy,
		holder,
		ConcurrencySlotTTL(input.Policy, input.DefaultTTL, input.TTLBuffer),
	)
	return slot, holder, err
}
