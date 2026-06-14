package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupPendingLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.CleanupPendingLease(ctx, leaseapp.CleanupPendingLeaseInput{
		Store:         c.deps.store,
		Limiter:       c.deps.providerConcurrency,
		Locks:         c.deps.locks,
		DataPlane:     c.deps.dataPlane,
		LocalProtocol: c.deps.cfg.LocalProtocol,
		Lease:         lease,
		IsNotFound:    isStoreNotFound,
		ReleaseProvider: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			return c.releaseLeaseProviderSession(ctx, lease)
		},
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	})
}
