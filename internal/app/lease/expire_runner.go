package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ExpireLeaseRunner struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	Clock                             Clock
	LocalProtocol                     string
	IsNotFound                        StoreNotFoundFunc
	ResolveGatewaysForLease           ProviderSessionGatewaysResolverFactory
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func (r ExpireLeaseRunner) Expire(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return ExpireLease(ctx, ExpireLeaseInput{
		Store:                             r.Store,
		Limiter:                           r.Limiter,
		Locks:                             r.Locks,
		DataPlane:                         r.DataPlane,
		Factory:                           r.Factory,
		LocalProtocol:                     r.LocalProtocol,
		Lease:                             lease,
		IsNotFound:                        r.IsNotFound,
		Now:                               r.now().UTC(),
		ResolveGateways:                   r.resolveGateways(lease),
		ObserveFinalConcurrencyReleaseErr: r.ObserveFinalConcurrencyReleaseErr,
	})
}

func (r ExpireLeaseRunner) resolveGateways(lease *proxyruntimev1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	if r.ResolveGatewaysForLease == nil {
		return nil
	}
	return r.ResolveGatewaysForLease(lease)
}

func (r ExpireLeaseRunner) now() time.Time {
	if r.Clock != nil {
		return r.Clock.Now()
	}
	return time.Now()
}
