package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

func runtimeLeaseCoordinatorDependencies(runtime *Runtime) leaseCoordinatorDependencies {
	if runtime == nil {
		return leaseCoordinatorDependencies{}
	}
	var store leaseapp.OrchestrationStore
	if runtime.store != nil {
		store = runtime.store
	}
	var locks leaseapp.LockManager
	if runtime.leaseLocks != nil {
		locks = leaseRuntimeLockManager{locks: runtime.leaseLocks}
	}
	return leaseCoordinatorDependencies{
		cfg:                     runtime.cfg,
		store:                   store,
		settings:                runtime.settings,
		locks:                   locks,
		dataPlane:               leaseRuntimeDataPlaneApplier{dataPlane: runtime.dataPlane, metrics: runtime.metrics},
		dynamicIPSelector:       runtime.dynamicIPSelector,
		sessionProviders:        leaseRegistrySessionProviderFactory{registry: runtime.accountProviders, client: runtime.providerHTTPClient, metrics: runtime.metrics},
		providerConcurrency:     runtime.providerConcurrency,
		logger:                  runtime.logger,
		exitCheckCache:          &runtime.exitCheckCache,
		leaseListener:           runtime.leaseListener,
		localListenerEndpoint:   runtime.localListenerEndpoint,
		sessionAdvertisedHost:   runtime.sessionAdvertisedHost,
		dynamicLeaseDialerProxy: runtime.dynamicLeaseDialerProxy,
		closeInUserConnections:  runtime.closeMihomoInUserConnections,
	}
}
