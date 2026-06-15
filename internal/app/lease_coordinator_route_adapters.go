package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

func (c leaseCoordinator) leaseRouteRetirer() leaseapp.LeaseRouteRetirer {
	return leaseRouteRetirerFactory{
		deps:             c.deps,
		settings:         c.settingsAdapter(),
		sideEffects:      c.routeSideEffects(),
		observeFinalSlot: warnFinalLeaseConcurrencyReleaseFailed(c.deps.logger),
	}.New()
}

func (c leaseCoordinator) releaseRunner() leaseapp.ReleaseRunner {
	return leaseReleaseRunnerFactory{
		deps:   c.deps,
		retire: c.leaseRouteRetirer(),
	}.New()
}

func (c leaseCoordinator) leaseRouteRestorer(settings *runtimeSettingsFile) leaseapp.LeaseRouteRestorer {
	return leaseRouteRestorerFactory{
		deps:     c.deps,
		settings: settings,
		adapter:  c.settingsAdapter(),
	}.New()
}
