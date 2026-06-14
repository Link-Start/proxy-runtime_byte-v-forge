package app

import (
	"context"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) applyAcquiredLeaseDataPlaneRoute(ctx context.Context, route leaseapp.SessionRoute, failure *leaseAcquireFailure) error {
	if err := leaseapp.UpsertSessionRoute(ctx, c.deps.dataPlane, route); err != nil {
		failure.afterRoute(route, "dataplane route apply failed")
		return unavailable("dataplane route apply failed", err)
	}
	return nil
}
