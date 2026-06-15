package app

import (
	"context"

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
		leases:    newRuntimeLeaseApplication(runtimeLeaseDependencies(runtime)),
		checks:    newRuntimeCheckApplication(runtimeCheckDependencies(runtime)),
		settings:  newRuntimeSettingsApplication(runtimeSettingsDependencies(runtime)),
		status:    newRuntimeStatusApplication(runtimeStatusDependencies(runtime)),
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
	mihomoNativeApply := newRuntimeMihomoNativeApplyScheduler(runtime)
	return runtimeSettingsApplicationDependencies{
		Logger:     runtime.logger,
		Settings:   runtime.settings,
		ProxyUsers: runtime.cfg.ProxyUsers,
		IPFraudProviderViews: func() []*proxyruntimev1.ProxyIPFraudProviderDescriptor {
			if runtime.ipFraudProviders == nil {
				return nil
			}
			return runtime.ipFraudProviders.ProviderDescriptors()
		},
		IPGeoProviderViews: func() []*proxyruntimev1.ProxyIPGeoProviderDescriptor {
			if runtime.ipGeoProviders == nil {
				return nil
			}
			return runtime.ipGeoProviders.ProviderDescriptors()
		},
		LoadMihomoNativeSettings: func(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
			if runtime.settings == nil {
				return normalizeMihomoNativeSettings(nil), nil
			}
			return runtime.settings.loadMihomoNative(ctx)
		},
		UpdateMihomoNativeSettings: func(ctx context.Context, config *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
			return updateMihomoNativeSettings(ctx, mihomoNativeUpdateDependencies{
				Repository: runtime.settings,
				ConfigDir:  runtime.cfg.Mihomo.ConfigDir,
				AfterApply: mihomoNativeApply.Schedule,
			}, config)
		},
		ScheduleApply: settingsApply.Schedule,
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
