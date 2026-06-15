package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var (
	ErrDynamicIPSelectorRequired              = errors.New("dynamic IP selector is required")
	ErrSelectedAcquireAttemptRunnerFactoryNil = errors.New("selected acquire attempt runner factory is required")
)

type DynamicIPSelectorFunc func(context.Context, *proxyruntimev1.AcquireProxyLeaseRequest) (DynamicIPSelection, error)

type SelectedAcquireAttemptRunnerFactory func(DynamicIPSelection) SelectedAcquireAttemptRunner

type AcquireAttemptErrorMapper func(error) error

type DynamicAcquireAttemptRunner struct {
	Select            DynamicIPSelectorFunc
	NewSelectedRunner SelectedAcquireAttemptRunnerFactory
	MapSelectionError AcquireAttemptErrorMapper
	MapAttemptError   AcquireAttemptErrorMapper
}

func (r DynamicAcquireAttemptRunner) Run(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, policy *proxyruntimev1.ProxySessionPolicy) (*proxyruntimev1.ProxyDynamicLease, error) {
	if r.Select == nil {
		return nil, ErrDynamicIPSelectorRequired
	}
	selection, err := r.Select(ctx, req)
	if err != nil {
		return nil, mapAcquireAttemptError(r.MapSelectionError, err)
	}
	if r.NewSelectedRunner == nil {
		return nil, ErrSelectedAcquireAttemptRunnerFactoryNil
	}
	runner := r.NewSelectedRunner(selection)
	lease, err := runner.Run(ctx, SelectedAcquireAttemptRunnerInput{
		SelectionPlan: selection.Plan,
		Policy:        policy,
	})
	if err != nil {
		return nil, mapAcquireAttemptError(r.MapAttemptError, err)
	}
	return lease, nil
}

func mapAcquireAttemptError(mapper AcquireAttemptErrorMapper, err error) error {
	if err == nil || mapper == nil {
		return err
	}
	return mapper(err)
}
