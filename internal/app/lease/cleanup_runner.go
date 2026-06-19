package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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

func (r CleanupPendingLeaseRunner) Cleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
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

func (r CleanupPendingLeaseRunner) resolveGateways(lease *proxyruntimev1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	if r.ResolveGatewaysForLease == nil {
		return nil
	}
	return r.ResolveGatewaysForLease(lease)
}
