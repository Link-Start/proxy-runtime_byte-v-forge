package lease

import (
	"context"
	"errors"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type CleanupPendingLeaseRunner struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	LocalProtocol                     string
	IsNotFound                        StoreNotFoundFunc
	ResolveGatewaysForLease           ProviderSessionGatewaysResolverFactory
	ObserveProviderReleaseFailure     LeaseErrorObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func (r CleanupPendingLeaseRunner) Cleanup(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	return CleanupPendingLease(ctx, CleanupPendingLeaseInput{
		Store:                             r.Store,
		Limiter:                           r.Limiter,
		Locks:                             r.Locks,
		DataPlane:                         r.DataPlane,
		Factory:                           r.Factory,
		LocalProtocol:                     r.LocalProtocol,
		Lease:                             lease,
		IsNotFound:                        r.IsNotFound,
		ResolveGateways:                   r.resolveGateways(lease),
		ObserveProviderReleaseFailure:     r.ObserveProviderReleaseFailure,
		ObserveFinalConcurrencyReleaseErr: r.ObserveFinalConcurrencyReleaseErr,
	})
}

func (r CleanupPendingLeaseRunner) resolveGateways(lease *proxygatewayv1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	if r.ResolveGatewaysForLease == nil {
		return nil
	}
	return r.ResolveGatewaysForLease(lease)
}

func MarkCleanupPending(lease *proxygatewayv1.ProxyDynamicLease, routePending bool, providerPending bool, finalStatus string) {
	if lease == nil {
		return
	}
	session := lease.GetSession()
	if session == nil {
		session = &proxygatewayv1.ProxySession{}
		lease.Session = session
	}
	if session.Labels == nil {
		session.Labels = map[string]string{}
	}
	if routePending {
		session.Labels[RouteCleanupPendingLabel] = "true"
	}
	if providerPending {
		session.Labels[ProviderCleanupPendingLabel] = "true"
	}
	if strings.TrimSpace(finalStatus) != "" {
		session.Labels[CleanupFinalStatusLabel] = strings.TrimSpace(finalStatus)
	}
}

func ClearCleanupPending(lease *proxygatewayv1.ProxyDynamicLease, routePending bool, providerPending bool) {
	if lease == nil || lease.GetSession() == nil || lease.GetSession().Labels == nil {
		return
	}
	if routePending {
		delete(lease.GetSession().Labels, RouteCleanupPendingLabel)
	}
	if providerPending {
		delete(lease.GetSession().Labels, ProviderCleanupPendingLabel)
	}
}

func MarkFailedAcquireCleanupPending(session *proxygatewayv1.ProxySession, routePending bool, providerPending bool) {
	lease := &proxygatewayv1.ProxyDynamicLease{Session: session}
	MarkCleanupPending(lease, routePending, providerPending, CleanupFinalFailed)
}

type CleanupPendingLeaseInput struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	LocalProtocol                     string
	Lease                             *proxygatewayv1.ProxyDynamicLease
	IsNotFound                        StoreNotFoundFunc
	ResolveGateways                   ProviderSessionGatewaysResolver
	ObserveProviderReleaseFailure     LeaseErrorObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func CleanupPendingLease(ctx context.Context, input CleanupPendingLeaseInput) error {
	return RunCurrentLeaseAction(ctx, CurrentLeaseActionInput{
		Store:      input.Store,
		Locks:      input.Locks,
		Lease:      input.Lease,
		IsNotFound: input.IsNotFound,
		Action: func(ctx context.Context, current *proxygatewayv1.ProxyDynamicLease) error {
			return cleanupCurrentPendingLease(ctx, input, current)
		},
	})
}

func cleanupCurrentPendingLease(ctx context.Context, input CleanupPendingLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	if !CleanupPending(lease) {
		return nil
	}
	if RouteCleanupPending(lease) {
		if err := cleanupPendingLeaseRoute(ctx, input, lease); err != nil {
			return err
		}
		ClearCleanupPending(lease, true, false)
	}
	if ProviderCleanupPending(lease) {
		if err := cleanupPendingProviderSession(ctx, input, lease); err != nil {
			return err
		}
		ClearCleanupPending(lease, false, true)
	}
	return saveCleanupProgressState(ctx, input, lease)
}

func cleanupPendingLeaseRoute(ctx context.Context, input CleanupPendingLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	return CleanupLeaseRoute(ctx, RouteCleanupInput{
		DataPlane:     input.DataPlane,
		Lease:         lease,
		LocalProtocol: input.LocalProtocol,
		RecordFailure: func(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
			return SaveCleanupRetry(ctx, input.Store, lease, "lease route cleanup failed")
		},
	})
}

func cleanupPendingProviderSession(ctx context.Context, input CleanupPendingLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	err := ReleaseLeaseProviderSessionWithLock(ctx, input.Locks, ProviderSessionReleaseInput{
		Store:           input.Store,
		Factory:         input.Factory,
		Lease:           lease,
		ResolveGateways: input.ResolveGateways,
		RecordFailure: func(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease, err error) error {
			observeLeaseErr(ctx, input.ObserveProviderReleaseFailure, lease, err)
			return SaveCleanupRetry(ctx, input.Store, lease, "provider session cleanup failed")
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func saveCleanupProgressState(ctx context.Context, input CleanupPendingLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	err := SaveCleanupProgress(ctx, input.Store, input.Limiter, lease)
	if errors.Is(err, ErrFinalLeaseConcurrencyRelease) {
		observeLease(ctx, input.ObserveFinalConcurrencyReleaseErr, lease)
		return nil
	}
	return err
}
