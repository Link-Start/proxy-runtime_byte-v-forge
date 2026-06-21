package lease

import (
	"context"
	"errors"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

var ErrSelectedAttemptProviderRunnerFactoryRequired = errors.New("selected attempt provider runner factory is required")

type SelectedAttemptProviderRunnerFactory func(SelectedAcquireAttempt) ProviderAccountAcquireRunner

type SelectedAttemptProviderAccountAction struct {
	Selection DynamicIPSelection
	Request   *proxygatewayv1.AcquireProxyLeaseRequest
	NewRunner SelectedAttemptProviderRunnerFactory
	MapError  ProviderAccountAcquireApplyErrorMapper
}

func (a SelectedAttemptProviderAccountAction) Run(ctx context.Context, attempt SelectedAcquireAttempt) (*proxygatewayv1.ProxyDynamicLease, error) {
	if a.NewRunner == nil {
		return nil, ErrSelectedAttemptProviderRunnerFactoryRequired
	}
	runner := a.NewRunner(attempt)
	lease, err := runner.Acquire(ctx, ProviderAccountAcquireRunInput{
		ProviderAccountID: attempt.ProviderAccountID,
		Gateway:           a.Selection.Endpoint,
		Request:           a.Request,
		SelectionPlan:     a.Selection.Plan,
		ConcurrencyHolder: attempt.ConcurrencyHolder,
	})
	if err != nil {
		return nil, a.mapError(err)
	}
	return lease, nil
}

func (a SelectedAttemptProviderAccountAction) mapError(err error) error {
	if err == nil || a.MapError == nil {
		return err
	}
	return a.MapError(err)
}
