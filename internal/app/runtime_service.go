package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type RuntimeService struct {
	proxyruntimev1.UnimplementedProxyRuntimeServiceServer
	providers runtimeProviderApplication
	leases    runtimeLeaseApplication
	checks    runtimeCheckApplication
	settings  runtimeSettingsApplication
	status    runtimeStatusApplication
}

var _ proxyruntimev1.ProxyRuntimeServiceServer = (*RuntimeService)(nil)

func NewRuntimeService(runtime *Runtime) *RuntimeService {
	return &RuntimeService{
		providers: newRuntimeProviderApplication(runtimeProviderDependencies(runtime)),
		leases:    newRuntimeLeaseApplication(runtime),
		checks:    newRuntimeCheckApplication(runtimeCheckDependencies(runtime)),
		settings:  newRuntimeSettingsApplication(runtime),
		status:    newRuntimeStatusApplication(runtime),
	}
}

func runtimeProviderDependencies(runtime *Runtime) runtimeProviderApplicationDependencies {
	if runtime == nil {
		return runtimeProviderApplicationDependencies{}
	}
	var locks leaseapp.LockManager
	if runtime.leaseLocks != nil {
		locks = leaseRuntimeLockManager{locks: runtime.leaseLocks}
	}
	var providerDescriptors runtimeProviderDescriptorsFunc
	if runtime.accountProviders != nil {
		providerDescriptors = runtime.accountProviders.Descriptors
	}
	return runtimeProviderApplicationDependencies{
		Store:               runtime.store,
		Settings:            runtime.settings,
		ProviderDescriptors: providerDescriptors,
		Locks:               locks,
		LeaseOperations: func() runtimeProviderLeaseOperations {
			return runtime.service().leases
		},
		Logger: runtime.logger,
	}
}

func runtimeCheckDependencies(runtime *Runtime) runtimeCheckApplicationDependencies {
	if runtime == nil {
		return runtimeCheckApplicationDependencies{}
	}
	return runtimeCheckApplicationDependencies{
		Settings:       runtime.settings,
		CheckClient:    runtime.checkProxyHTTPClient,
		ProbeExitIP:    runtime.probeExitIP,
		LookupGeo:      runtime.lookupIPGeo,
		CheckFraud:     runtime.checkIPFraud,
		RunEdgeCanary:  runtime.runEdgeCanary,
		ExitCheckCache: &runtime.exitCheckCache,
	}
}

func (r *Runtime) service() *RuntimeService {
	if r.appService != nil {
		return r.appService
	}
	return NewRuntimeService(r)
}
