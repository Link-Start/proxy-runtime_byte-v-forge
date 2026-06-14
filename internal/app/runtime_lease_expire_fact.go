package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) expireLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.ExpireLease(ctx, leaseapp.ExpireLeaseInput{
		Store:                             c.deps.store,
		Limiter:                           c.deps.providerConcurrency,
		Locks:                             c.deps.locks,
		DataPlane:                         c.deps.dataPlane,
		Factory:                           c.deps.sessionProviders,
		LocalProtocol:                     c.deps.cfg.LocalProtocol,
		Lease:                             lease,
		IsNotFound:                        isStoreNotFound,
		Now:                               c.now().UTC(),
		ResolveGateways:                   c.providerSessionGatewaysResolver(lease),
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	})
}
