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

type SelectedAcquireAttemptAction func(context.Context, SelectedAcquireAttempt) (*proxyruntimev1.ProxyDynamicLease, error)

type SelectedAcquireAttemptLimitFunc func(*proxyruntimev1.ProxyDynamicIPSelectionPlan, *proxyruntimev1.ProxySessionPolicy) uint32

type SelectedAcquireAttemptRunner struct {
	Store          OrchestrationStore
	IDs            IDGenerator
	Limiter        ProviderAccountConcurrencyLimiter
	Locks          LockManager
	Limit          SelectedAcquireAttemptLimitFunc
	DefaultTTL     time.Duration
	TTLBuffer      time.Duration
	ReleaseTimeout time.Duration
	Action         SelectedAcquireAttemptAction
}

type SelectedAcquireAttemptRunnerInput struct {
	SelectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan
	Policy        *proxyruntimev1.ProxySessionPolicy
}

type SelectedAcquireAttemptRunInput struct {
	Store          OrchestrationStore
	IDs            IDGenerator
	Limiter        ProviderAccountConcurrencyLimiter
	Locks          LockManager
	SelectionPlan  *proxyruntimev1.ProxyDynamicIPSelectionPlan
	Limit          uint32
	Policy         *proxyruntimev1.ProxySessionPolicy
	DefaultTTL     time.Duration
	TTLBuffer      time.Duration
	ReleaseTimeout time.Duration
	Action         SelectedAcquireAttemptAction
}

func (r SelectedAcquireAttemptRunner) Run(ctx context.Context, input SelectedAcquireAttemptRunnerInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	return RunSelectedAcquireAttempt(ctx, SelectedAcquireAttemptRunInput{
		Store:          r.Store,
		IDs:            r.IDs,
		Limiter:        r.Limiter,
		Locks:          r.Locks,
		SelectionPlan:  input.SelectionPlan,
		Limit:          r.limit(input.SelectionPlan, input.Policy),
		Policy:         input.Policy,
		DefaultTTL:     r.DefaultTTL,
		TTLBuffer:      r.TTLBuffer,
		ReleaseTimeout: r.ReleaseTimeout,
		Action:         r.Action,
	})
}

func (r SelectedAcquireAttemptRunner) limit(selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	if r.Limit == nil {
		return 0
	}
	return r.Limit(selectionPlan, policy)
}

func RunSelectedAcquireAttempt(ctx context.Context, input SelectedAcquireAttemptRunInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	attempt, err := PrepareSelectedAcquireAttempt(ctx, SelectedAcquireAttemptInput{
		Store:         input.Store,
		IDs:           input.IDs,
		Limiter:       input.Limiter,
		SelectionPlan: input.SelectionPlan,
		Limit:         input.Limit,
		Policy:        input.Policy,
		DefaultTTL:    input.DefaultTTL,
		TTLBuffer:     input.TTLBuffer,
	})
	if err != nil {
		return nil, err
	}
	if input.Action == nil {
		return nil, ErrAcquireAttemptActionRequired
	}
	return RunLockedAcquireAttempt(ctx, LockedAcquireAttemptInput{
		Locks:             input.Locks,
		ProviderAccountID: attempt.ProviderAccountID,
		ConcurrencySlot:   attempt.ConcurrencySlot,
		ReleaseTimeout:    input.ReleaseTimeout,
		Action: func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
			return input.Action(ctx, attempt)
		},
	})
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
