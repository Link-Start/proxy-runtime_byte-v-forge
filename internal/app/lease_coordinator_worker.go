package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

func (c leaseCoordinator) workerProcessor() leaseapp.WorkerProcessor {
	return leaseWorkerProcessorFactory{
		deps:    c.deps,
		restore: c.restoreLeaseRunner(),
		expire:  c.expireLeaseRunner(),
		cleanup: c.cleanupPendingLeaseRunner(),
	}.New()
}

func (c leaseCoordinator) restoreLeaseRunner() leaseapp.RestoreLeaseRouteRunner {
	return leaseRestoreRunnerFactory{deps: c.deps}.New()
}

func (c leaseCoordinator) expireLeaseRunner() leaseapp.ExpireLeaseRunner {
	return leaseExpireRunnerFactory{
		deps:     c.deps,
		settings: c.settingsAdapter(),
	}.New()
}

func (c leaseCoordinator) cleanupPendingLeaseRunner() leaseapp.CleanupPendingLeaseRunner {
	return leaseCleanupPendingRunnerFactory{
		deps:     c.deps,
		settings: c.settingsAdapter(),
	}.New()
}
