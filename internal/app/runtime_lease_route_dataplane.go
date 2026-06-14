package app

import (
	"context"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) applyAcquiredLeaseDataPlaneRoute(ctx context.Context, route leaseapp.SessionRoute, failure *leaseapp.FailedAcquireRecorder) error {
	if err := leaseapp.UpsertSessionRoute(ctx, c.deps.dataPlane, route); err != nil {
		failure.AfterRoute(ctx, route, "dataplane route apply failed")
		return unavailable("dataplane route apply failed", err)
	}
	return nil
}
