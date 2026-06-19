package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

var ErrProviderAccountAcquireApplyRequired = errors.New("provider account acquire apply action is required")

type ProviderAccountAcquireInput struct {
	Store              OrchestrationStore
	IDs                IDGenerator
	Clock              clock.Clock
	DataPlane          DataPlaneApplier
	Logger             Logger
	Factory            SessionProviderFactory
	Locks              LockManager
	ProviderAccountID  string
	Gateway            accountproxy.Gateway
	Request            *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan      *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ConcurrencyHolder  string
	ResolveLineBinding RouteLineBindingResolver
	Apply              ProviderAccountAcquireApply
}

type ProviderAccountAcquireRunner struct {
	Store              OrchestrationStore
	IDs                IDGenerator
	Clock              clock.Clock
	DataPlane          DataPlaneApplier
	Logger             Logger
	Factory            SessionProviderFactory
	Locks              LockManager
	ResolveLineBinding RouteLineBindingResolver
	Apply              ProviderAccountAcquireApply
}

type ProviderAccountAcquireRunInput struct {
	ProviderAccountID string
	Gateway           accountproxy.Gateway
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ConcurrencyHolder string
}

type ProviderAccountAcquireApplyInput struct {
	ProviderAccountID string
	ProviderClient    SessionProvider
	Session           *proxyruntimev1.ProxySession
	Nodes             []provider.Node
	DialerProxy       string
	LineLabels        map[string]string
	Failure           *FailedAcquireRecorder
}

type ProviderAccountAcquireApply func(context.Context, ProviderAccountAcquireApplyInput) (*proxyruntimev1.ProxyDynamicLease, error)

func (r ProviderAccountAcquireRunner) Acquire(ctx context.Context, input ProviderAccountAcquireRunInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	return RunProviderAccountAcquire(ctx, ProviderAccountAcquireInput{
		Store:              r.Store,
		IDs:                r.IDs,
		Clock:              r.Clock,
		DataPlane:          r.DataPlane,
		Logger:             r.Logger,
		Factory:            r.Factory,
		Locks:              r.Locks,
		ProviderAccountID:  input.ProviderAccountID,
		Gateway:            input.Gateway,
		Request:            input.Request,
		SelectionPlan:      input.SelectionPlan,
		ConcurrencyHolder:  input.ConcurrencyHolder,
		ResolveLineBinding: r.ResolveLineBinding,
		Apply:              r.Apply,
	})
}

func RunProviderAccountAcquire(ctx context.Context, input ProviderAccountAcquireInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	providerSession, err := AcquireProviderSession(ctx, ProviderSessionAcquireInput{
		Store:             input.Store,
		Factory:           input.Factory,
		ProviderAccountID: input.ProviderAccountID,
		Gateway:           input.Gateway,
		Request:           input.Request,
		SelectionPlan:     input.SelectionPlan,
		ConcurrencyHolder: input.ConcurrencyHolder,
	})
	if err != nil {
		if errors.Is(err, ErrProviderSessionFetch) {
			failure := newProviderAccountAcquireFailure(input, providerSession)
			failure.BeforeRoute(ctx, "provider session fetch failed")
		}
		return nil, err
	}
	failure := newProviderAccountAcquireFailure(input, providerSession)
	lineBinding, err := PrepareRouteLineBinding(ctx, providerSession.Nodes, input.Request.GetAccountId(), input.ResolveLineBinding)
	if err != nil {
		failure.BeforeRoute(ctx, "lease line resolution failed")
		return nil, err
	}
	if input.Apply == nil {
		return nil, ErrProviderAccountAcquireApplyRequired
	}
	return RunSessionListenerAllocation(ctx, input.Locks, func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
		return input.Apply(ctx, ProviderAccountAcquireApplyInput{
			ProviderAccountID: providerSession.ProviderAccountID,
			ProviderClient:    providerSession.ProviderClient,
			Session:           providerSession.Session,
			Nodes:             lineBinding.Nodes,
			DialerProxy:       lineBinding.DialerProxy,
			LineLabels:        lineBinding.Labels,
			Failure:           failure,
		})
	})
}

func newProviderAccountAcquireFailure(input ProviderAccountAcquireInput, providerSession ProviderSessionAcquireResult) *FailedAcquireRecorder {
	return NewFailedAcquireRecorder(FailedAcquireRecorderInput{
		Store:             input.Store,
		IDs:               input.IDs,
		Clock:             input.Clock,
		DataPlane:         input.DataPlane,
		Logger:            input.Logger,
		Request:           input.Request,
		ProviderAccountID: providerSession.ProviderAccountID,
		ProviderClient:    providerSession.ProviderClient,
		Session:           providerSession.Session,
		SelectionPlan:     input.SelectionPlan,
	})
}
