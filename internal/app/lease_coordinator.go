package app

type leaseCoordinator struct {
	runtime *Runtime
}

func newLeaseCoordinator(runtime *Runtime) leaseCoordinator {
	return leaseCoordinator{runtime: runtime}
}
