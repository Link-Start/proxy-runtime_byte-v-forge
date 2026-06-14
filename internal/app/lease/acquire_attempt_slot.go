package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type AcquireAttemptSlotStage int

const (
	AcquireAttemptSlotNoError AcquireAttemptSlotStage = iota
	AcquireAttemptSlotProviderAccountLookup
	AcquireAttemptSlotConcurrencyAcquire
)

type AcquireAttemptSlotInput struct {
	Store             OrchestrationStore
	Limiter           ProviderAccountConcurrencyLimiter
	ProviderAccountID string
	Limit             uint32
	Policy            *proxyruntimev1.ProxySessionPolicy
	LeaseID           string
	DefaultTTL        time.Duration
	TTLBuffer         time.Duration
}

type AcquireAttemptSlot struct {
	ConcurrencySlot   ProviderAccountConcurrencySlot
	ConcurrencyHolder string
}

func AcquireAttemptSlotForProviderAccount(ctx context.Context, input AcquireAttemptSlotInput) (AcquireAttemptSlot, AcquireAttemptSlotStage, error) {
	providerAccount, err := input.Store.ProviderAccount(ctx, input.ProviderAccountID)
	if err != nil {
		return AcquireAttemptSlot{}, AcquireAttemptSlotProviderAccountLookup, err
	}
	slot, holder, err := AcquireAttemptConcurrencySlot(ctx, AcquireAttemptConcurrencyInput{
		Limiter:    input.Limiter,
		AccountID:  providerAccount.GetAccountId(),
		Limit:      input.Limit,
		Policy:     input.Policy,
		LeaseID:    input.LeaseID,
		DefaultTTL: input.DefaultTTL,
		TTLBuffer:  input.TTLBuffer,
	})
	if err != nil {
		return AcquireAttemptSlot{}, AcquireAttemptSlotConcurrencyAcquire, err
	}
	return AcquireAttemptSlot{ConcurrencySlot: slot, ConcurrencyHolder: holder}, AcquireAttemptSlotNoError, nil
}
