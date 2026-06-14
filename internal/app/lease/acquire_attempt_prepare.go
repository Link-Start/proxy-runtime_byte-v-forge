package lease

import (
	"context"
	"errors"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAcquireAttemptLeaseID = errors.New("generate lease id")

type SelectedAcquireAttemptInput struct {
	Store         OrchestrationStore
	IDs           IDGenerator
	Limiter       ProviderAccountConcurrencyLimiter
	SelectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan
	Limit         uint32
	Policy        *proxyruntimev1.ProxySessionPolicy
	DefaultTTL    time.Duration
	TTLBuffer     time.Duration
}

type SelectedAcquireAttempt struct {
	ProviderAccountID string
	LeaseID           string
	ConcurrencySlot   ProviderAccountConcurrencySlot
	ConcurrencyHolder string
}

func PrepareSelectedAcquireAttempt(ctx context.Context, input SelectedAcquireAttemptInput) (SelectedAcquireAttempt, error) {
	providerAccountID := SelectedProviderAccountID(input.SelectionPlan)
	leaseID, err := newAttemptLeaseID(input.IDs)
	if err != nil {
		return SelectedAcquireAttempt{}, err
	}
	attemptSlot, err := AcquireAttemptSlotForProviderAccount(ctx, AcquireAttemptSlotInput{
		Store:             input.Store,
		Limiter:           input.Limiter,
		ProviderAccountID: providerAccountID,
		Limit:             input.Limit,
		Policy:            input.Policy,
		LeaseID:           leaseID,
		DefaultTTL:        input.DefaultTTL,
		TTLBuffer:         input.TTLBuffer,
	})
	if err != nil {
		return SelectedAcquireAttempt{}, err
	}
	return SelectedAcquireAttempt{
		ProviderAccountID: providerAccountID,
		LeaseID:           leaseID,
		ConcurrencySlot:   attemptSlot.ConcurrencySlot,
		ConcurrencyHolder: attemptSlot.ConcurrencyHolder,
	}, nil
}

func newAttemptLeaseID(ids IDGenerator) (string, error) {
	if ids == nil {
		return "", fmt.Errorf("%w: lease id generator is required", ErrAcquireAttemptLeaseID)
	}
	leaseID, err := ids.NewLeaseID()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAcquireAttemptLeaseID, err)
	}
	return leaseID, nil
}
