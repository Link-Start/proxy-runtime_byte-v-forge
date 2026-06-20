package lease

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

var ErrProviderAccountAcquireApplyRequired = errors.New("provider account acquire apply action is required")

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
	providerSession, err := AcquireProviderSession(ctx, ProviderSessionAcquireInput{
		Store:             r.Store,
		Factory:           r.Factory,
		ProviderAccountID: input.ProviderAccountID,
		Gateway:           input.Gateway,
		Request:           input.Request,
		SelectionPlan:     input.SelectionPlan,
		ConcurrencyHolder: input.ConcurrencyHolder,
	})
	if err != nil {
		if errors.Is(err, ErrProviderSessionFetch) {
			failure := r.newFailure(input.Request, input.SelectionPlan, providerSession)
			failure.BeforeRoute(ctx, "provider session fetch failed")
		}
		return nil, err
	}
	failure := r.newFailure(input.Request, input.SelectionPlan, providerSession)
	lineBinding, err := PrepareRouteLineBinding(ctx, providerSession.Nodes, input.Request.GetAccountId(), r.ResolveLineBinding)
	if err != nil {
		failure.BeforeRoute(ctx, "lease line resolution failed")
		return nil, err
	}
	if r.Apply == nil {
		return nil, ErrProviderAccountAcquireApplyRequired
	}
	return RunSessionListenerAllocation(ctx, r.Locks, func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
		return r.Apply(ctx, ProviderAccountAcquireApplyInput{
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

func (r ProviderAccountAcquireRunner) newFailure(request *proxyruntimev1.AcquireProxyLeaseRequest, selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, providerSession ProviderSessionAcquireResult) *FailedAcquireRecorder {
	return NewFailedAcquireRecorder(FailedAcquireRecorderInput{
		Store:             r.Store,
		IDs:               r.IDs,
		Clock:             r.Clock,
		DataPlane:         r.DataPlane,
		Logger:            r.Logger,
		Request:           request,
		ProviderAccountID: providerSession.ProviderAccountID,
		ProviderClient:    providerSession.ProviderClient,
		Session:           providerSession.Session,
		SelectionPlan:     selectionPlan,
	})
}

type ProviderAccountAcquireApplyErrorMapper func(error) error

type ProviderAccountAcquiredRouteApplier struct {
	Applier           AcquiredRouteApplier
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ConcurrencyHolder string
	MapError          ProviderAccountAcquireApplyErrorMapper
}

func (a ProviderAccountAcquiredRouteApplier) Apply(ctx context.Context, acquired ProviderAccountAcquireApplyInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := a.Applier.Apply(ctx, AcquiredRouteApplierInput{
		Failure:           acquired.Failure,
		LeaseID:           a.LeaseID,
		Request:           a.Request,
		ProviderClient:    acquired.ProviderClient,
		ProviderAccountID: acquired.ProviderAccountID,
		ConcurrencyHolder: a.ConcurrencyHolder,
		Session:           acquired.Session,
		Nodes:             acquired.Nodes,
		DialerProxy:       acquired.DialerProxy,
		LineLabels:        acquired.LineLabels,
		SelectionPlan:     a.SelectionPlan,
	})
	if err != nil {
		return nil, a.mapError(err)
	}
	return lease, nil
}

func (a ProviderAccountAcquiredRouteApplier) mapError(err error) error {
	if err == nil || a.MapError == nil {
		return err
	}
	return a.MapError(err)
}

type ProviderAccountConcurrencyLimiter interface {
	Acquire(context.Context, string, *proxyruntimev1.ProxySessionPolicy, uint32, string, time.Duration) (ProviderAccountConcurrencySlot, error)
	Available(context.Context, string, *proxyruntimev1.ProxySessionPolicy, uint32, string) (bool, error)
	Release(context.Context, string, *proxyruntimev1.ProxySessionPolicy, string) error
}

func AcquireProviderAccountConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, accountID string, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (ProviderAccountConcurrencySlot, error) {
	accountID = strings.TrimSpace(accountID)
	if limiter == nil {
		return NoopProviderAccountConcurrencySlot{}, nil
	}
	slot, err := limiter.Acquire(ctx, accountID, policy, limit, holder, ttl)
	if err != nil {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached: %w", accountID, ConcurrencyModeText(policy), err)
	}
	return slot, nil
}

type ProviderAccountConcurrencySlot interface {
	Release(context.Context) error
}

type NoopProviderAccountConcurrencySlot struct{}

func (NoopProviderAccountConcurrencySlot) Release(context.Context) error {
	return nil
}

func ReleaseProviderAccountConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, accountID string, policy *proxyruntimev1.ProxySessionPolicy, holder string) error {
	accountID = strings.TrimSpace(accountID)
	holder = strings.TrimSpace(holder)
	if accountID == "" || holder == "" || limiter == nil {
		return nil
	}
	return limiter.Release(ctx, accountID, policy, holder)
}

func ReleaseLeaseConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	return ReleaseProviderAccountConcurrencySlot(ctx, limiter, lease.GetProviderAccountId(), ConcurrencyPolicy(lease), ConcurrencyHolder(lease))
}

func ReleaseConcurrencySlotUnlessKept(ctx context.Context, slot ProviderAccountConcurrencySlot, keep bool) error {
	if keep || slot == nil {
		return nil
	}
	return slot.Release(ctx)
}
