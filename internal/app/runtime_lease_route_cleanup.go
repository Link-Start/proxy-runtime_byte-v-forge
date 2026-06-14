package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, recordFailure leaseapp.RouteCleanupFailureRecorder) error {
	return leaseapp.CleanupLeaseRoute(ctx, leaseapp.RouteCleanupInput{
		DataPlane:     c.deps.dataPlane,
		Lease:         lease,
		LocalProtocol: c.deps.cfg.LocalProtocol,
		RecordFailure: recordFailure,
	})
}
