package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	providerapp "github.com/byte-v-forge/proxy-runtime/internal/app/provider/application"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings"
)

type RuntimeService struct {
	proxyruntimev1.UnimplementedProxyRuntimeServiceServer
	providers providerapp.Service
	leases    runtimeLeaseApplication
	checks    runtimeCheckApplication
	settings  settingsapp.Application
	status    runtimeStatusApplication
	metrics   *runtimeMetrics
	metricsUI runtimeMetricsApplication
}

var _ proxyruntimev1.ProxyRuntimeServiceServer = (*RuntimeService)(nil)

func NewRuntimeService(runtime *Runtime) *RuntimeService {
	return &RuntimeService{
		providers: providerapp.New(runtimeProviderDependencies(runtime)),
		leases:    newRuntimeLeaseApplication(runtimeLeaseDependencies(runtime)),
		checks:    newRuntimeCheckApplication(runtimeCheckDependencies(runtime)),
		settings:  newRuntimeSettingsApplication(runtimeSettingsDependencies(runtime)),
		status:    newRuntimeStatusApplication(runtimeStatusDependencies(runtime)),
		metrics:   runtimeMetricsFromRuntime(runtime),
		metricsUI: newRuntimeMetricsApplication(runtimeMetricsDependencies(runtime)),
	}
}

func runtimeMetricsFromRuntime(runtime *Runtime) *runtimeMetrics {
	if runtime == nil {
		return nil
	}
	return runtime.metrics
}

func runtimeMetricsDependencies(runtime *Runtime) runtimeMetricsApplicationDependencies {
	return runtimeMetricsApplicationDependencies{
		Metrics: runtimeMetricsFromRuntime(runtime),
	}
}

func runtimeProviderDependencies(runtime *Runtime) providerapp.Dependencies {
	if runtime == nil {
		return providerapp.Dependencies{}
	}
	var locks leaseapp.LockManager
	if runtime.leaseLocks != nil {
		locks = leaseRuntimeLockManager{locks: runtime.leaseLocks}
	}
	var providerDescriptors providerapp.DescriptorsFunc
	if runtime.accountProviders != nil {
		providerDescriptors = runtime.accountProviders.Descriptors
	}
	return providerapp.Dependencies{
		Store:               runtime.store,
		LoadSettings:        runtime.settings.load,
		ProviderDescriptors: providerDescriptors,
		Locks:               locks,
		LeaseOperations: func() providerapp.LeaseOperations {
			return runtime.service().leases
		},
		Logger: runtime.logger,
	}
}

func runtimeLeaseDependencies(runtime *Runtime) leaseapp.Dependencies {
	if runtime == nil {
		return leaseapp.Dependencies{}
	}
	return leaseapp.Dependencies{
		Repository:  runtime.store,
		Coordinator: runtime.leaseCoordinator,
		Worker:      runtime.leaseCoordinator.workerProcessor(),
		Logger:      runtime.logger,
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

func runtimeSettingsDependencies(runtime *Runtime) runtimeSettingsApplicationDependencies {
	if runtime == nil {
		return runtimeSettingsApplicationDependencies{}
	}
	settingsApply := newRuntimeSettingsApplyScheduler(runtime)
	providerViews := newRuntimeSettingsProviderViewAdapter(runtime)
	mihomoNative := newRuntimeSettingsMihomoNativeAdapter(runtime)
	return runtimeSettingsApplicationDependencies{
		Logger:                     runtime.logger,
		Settings:                   runtime.settings,
		ProxyUsers:                 runtime.cfg.ProxyUsers,
		IPFraudProviderViews:       providerViews.IPFraudProviderViews,
		IPGeoProviderViews:         providerViews.IPGeoProviderViews,
		LoadMihomoNativeSettings:   mihomoNative.Load,
		UpdateMihomoNativeSettings: mihomoNative.Update,
		ScheduleApply:              settingsApply.Schedule,
	}
}

func runtimeStatusDependencies(runtime *Runtime) runtimeStatusApplicationDependencies {
	if runtime == nil {
		return runtimeStatusApplicationDependencies{}
	}
	return runtimeStatusApplicationDependencies{
		RuntimeStatus: runtime.runtimeStatus,
	}
}

func (r *Runtime) service() *RuntimeService {
	if r.appService != nil {
		return r.appService
	}
	return NewRuntimeService(r)
}
