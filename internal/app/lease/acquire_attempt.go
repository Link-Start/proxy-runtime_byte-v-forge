package lease

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAcquireAttemptRunnerRequired = errors.New("acquire attempt runner is required")

type AcquireAttemptRunner func(attempt int) (*proxyruntimev1.ProxyDynamicLease, error)

type AcquireAttemptRetryPolicy func(error) bool

type AcquireAttemptFailureObserver func(attempt int, err error)

const (
	acquireRetryBaseBackoff = 100 * time.Millisecond
	acquireRetryMaxBackoff  = 2 * time.Second
)

func RunAcquireAttempts(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPolicy *proxyruntimev1.ProxyDynamicIPSelectionPolicy, run AcquireAttemptRunner, retry AcquireAttemptRetryPolicy, observe AcquireAttemptFailureObserver) (*proxyruntimev1.ProxyDynamicLease, error) {
	if run == nil {
		return nil, ErrAcquireAttemptRunnerRequired
	}
	maxAttempts := DynamicIPSelectionMaxAttempts(selectionPolicy)
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 1 {
			if err := sleepAcquireBackoff(ctx, attempt); err != nil {
				return nil, err
			}
		}
		SetAttemptLabel(req, attempt)
		lease, err := run(attempt)
		if err == nil {
			return lease, nil
		}
		lastErr = err
		if retry == nil || !retry(err) {
			return nil, err
		}
		if observe != nil {
			observe(attempt, err)
		}
	}
	return nil, lastErr
}

func sleepAcquireBackoff(ctx context.Context, attempt int) error {
	shift := attempt - 2
	if shift < 0 {
		shift = 0
	}
	if shift > 16 {
		shift = 16
	}
	backoff := acquireRetryBaseBackoff << shift
	if backoff <= 0 || backoff > acquireRetryMaxBackoff {
		backoff = acquireRetryMaxBackoff
	}
	timer := time.NewTimer(rand.N(backoff))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

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

func (r SelectedAcquireAttemptRunner) Run(ctx context.Context, input SelectedAcquireAttemptRunnerInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	attempt, err := PrepareSelectedAcquireAttempt(ctx, SelectedAcquireAttemptInput{
		Store:         r.Store,
		IDs:           r.IDs,
		Limiter:       r.Limiter,
		SelectionPlan: input.SelectionPlan,
		Limit:         r.limit(input.SelectionPlan, input.Policy),
		Policy:        input.Policy,
		DefaultTTL:    r.DefaultTTL,
		TTLBuffer:     r.TTLBuffer,
	})
	if err != nil {
		return nil, err
	}
	if r.Action == nil {
		return nil, ErrAcquireAttemptActionRequired
	}
	return RunLockedAcquireAttempt(ctx, LockedAcquireAttemptInput{
		Locks:             r.Locks,
		ProviderAccountID: attempt.ProviderAccountID,
		ConcurrencySlot:   attempt.ConcurrencySlot,
		ReleaseTimeout:    r.ReleaseTimeout,
		Action: func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
			return r.Action(ctx, attempt)
		},
	})
}

func (r SelectedAcquireAttemptRunner) limit(selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	if r.Limit == nil {
		return 0
	}
	return r.Limit(selectionPlan, policy)
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

var ErrAcquireAttemptActionRequired = errors.New("acquire attempt action is required")

type AcquireAttemptAction func(context.Context) (*proxyruntimev1.ProxyDynamicLease, error)

type LockedAcquireAttemptInput struct {
	Locks             LockManager
	ProviderAccountID string
	ConcurrencySlot   ProviderAccountConcurrencySlot
	ReleaseTimeout    time.Duration
	Action            AcquireAttemptAction
}

func RunLockedAcquireAttempt(ctx context.Context, input LockedAcquireAttemptInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	if input.Action == nil {
		return nil, ErrAcquireAttemptActionRequired
	}
	keepConcurrencySlot := false
	defer func() {
		releaseCtx := context.WithoutCancel(ctx)
		if input.ReleaseTimeout > 0 {
			var cancel context.CancelFunc
			releaseCtx, cancel = context.WithTimeout(releaseCtx, input.ReleaseTimeout)
			defer cancel()
		}
		_ = ReleaseConcurrencySlotUnlessKept(releaseCtx, input.ConcurrencySlot, keepConcurrencySlot)
	}()
	var lease *proxyruntimev1.ProxyDynamicLease
	err := WithProviderAccountLock(ctx, input.Locks, input.ProviderAccountID, func(ctx context.Context) error {
		var err error
		lease, err = input.Action(ctx)
		if err == nil {
			keepConcurrencySlot = true
		}
		return err
	})
	return lease, err
}

var ErrAcquireAttemptConcurrencyLimit = errors.New("provider account concurrency limit reached")

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

func AcquireAttemptSlotForProviderAccount(ctx context.Context, input AcquireAttemptSlotInput) (AcquireAttemptSlot, error) {
	providerAccount, err := input.Store.ProviderAccount(ctx, input.ProviderAccountID)
	if err != nil {
		return AcquireAttemptSlot{}, err
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
		return AcquireAttemptSlot{}, fmt.Errorf("%w: %w", ErrAcquireAttemptConcurrencyLimit, err)
	}
	return AcquireAttemptSlot{ConcurrencySlot: slot, ConcurrencyHolder: holder}, nil
}

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
