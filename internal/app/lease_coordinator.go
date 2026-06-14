package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

type leaseCoordinator struct {
	deps leaseCoordinatorDependencies
}

func newLeaseCoordinator(deps leaseCoordinatorDependencies) leaseCoordinator {
	if deps.clock == nil {
		deps.clock = leaseapp.SystemClock{}
	}
	if deps.ids == nil {
		deps.ids = randomLeaseIDGenerator{byteLength: leaseIDByteLength}
	}
	return leaseCoordinator{deps: deps}
}
