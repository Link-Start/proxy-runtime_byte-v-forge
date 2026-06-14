package lease

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
	return RunAcquireAttempts(input.Request, selectionPolicy, input.RunAttempt, input.Retry, input.Observe)
}

func IsAcquirePolicyError(err error) bool {
	return errors.Is(err, ErrProfileDynamicIPNotConfigured) ||
		errors.Is(err, ErrProfileLeaseRequiresSticky) ||
		errors.Is(err, ErrRequestRequiresStickyDynamicIP)
}
