package lease

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
)

type AccountLockedAcquireInput struct {
	Store               OrchestrationStore
	Request             *proxyruntimev1.AcquireProxyLeaseRequest
	EgressProfiles      []*proxyruntimev1.EgressProfileSettings
	Now                 time.Time
	PlaygroundAccountID string
	PlaygroundUsername  string
	Reuse               ExistingActiveLeaseAction
	Replace             ExistingActiveLeaseAction
	RunAttempt          AcquireAttemptRunner
	Retry               AcquireAttemptRetryPolicy
	Observe             AcquireAttemptFailureObserver
}

type AccountLockedAcquireRunner struct {
	Store               OrchestrationStore
	Clock               clock.Clock
	PlaygroundAccountID string
	PlaygroundUsername  string
	Reuse               ExistingActiveLeaseAction
	Replace             ExistingActiveLeaseAction
	RunAttempt          AcquireAttemptRunner
	Retry               AcquireAttemptRetryPolicy
	Observe             AcquireAttemptFailureObserver
}

type AccountLockedAcquireRunnerInput struct {
	Request        *proxyruntimev1.AcquireProxyLeaseRequest
	EgressProfiles []*proxyruntimev1.EgressProfileSettings
}

func (r AccountLockedAcquireRunner) Run(ctx context.Context, input AccountLockedAcquireRunnerInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	return RunAccountLockedAcquire(ctx, AccountLockedAcquireInput{
		Store:               r.Store,
		Request:             input.Request,
		EgressProfiles:      input.EgressProfiles,
		Now:                 r.now().UTC(),
		PlaygroundAccountID: r.PlaygroundAccountID,
		PlaygroundUsername:  r.PlaygroundUsername,
		Reuse:               r.Reuse,
		Replace:             r.Replace,
		RunAttempt:          r.RunAttempt,
		Retry:               r.Retry,
		Observe:             r.Observe,
	})
}

func (r AccountLockedAcquireRunner) now() time.Time {
	if r.Clock != nil {
		return r.Clock.Now()
	}
	return time.Now()
}

func RunAccountLockedAcquire(ctx context.Context, input AccountLockedAcquireInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	selectionPolicy, err := ApplyAcquireRequestPolicies(input.Request, input.EgressProfiles)
	if err != nil {
		return nil, err
	}
	existing, handled, err := HandleExistingActiveLease(ctx, ExistingActiveLeaseInput{
		Store:               input.Store,
		Request:             input.Request,
		Now:                 input.Now,
		PlaygroundAccountID: input.PlaygroundAccountID,
		PlaygroundUsername:  input.PlaygroundUsername,
		Reuse:               input.Reuse,
		Replace:             input.Replace,
	})
	if err != nil {
		return nil, err
	}
	if handled {
		return existing, nil
	}
	return RunAcquireAttempts(ctx, input.Request, selectionPolicy, input.RunAttempt, input.Retry, input.Observe)
}

func IsAcquirePolicyError(err error) bool {
	return errors.Is(err, ErrProfileDynamicIPNotConfigured) ||
		errors.Is(err, ErrProfileLeaseRequiresSticky) ||
		errors.Is(err, ErrRequestRequiresStickyDynamicIP)
}
