package app

import "github.com/byte-v-forge/proxy-runtime/internal/clock"

type leaseCoordinator struct {
	deps leaseCoordinatorDependencies
}

func newLeaseCoordinator(deps leaseCoordinatorDependencies) leaseCoordinator {
	if deps.clock == nil {
		deps.clock = clock.SystemClock{}
	}
	if deps.ids == nil {
		deps.ids = randomLeaseIDGenerator{byteLength: leaseIDByteLength}
	}
	return leaseCoordinator{deps: deps}
}

func (c leaseCoordinator) routeSideEffects() leaseRouteSideEffects {
	return leaseRouteSideEffects{
		exitCheckCache:         c.deps.exitCheckCache,
		closeInUserConnections: c.deps.closeInUserConnections,
	}
}
