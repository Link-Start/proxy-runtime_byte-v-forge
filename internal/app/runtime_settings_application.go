package app

import (
	"context"
	"log/slog"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type runtimeSettingsRepository interface {
	view(context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error)
	load(context.Context) (*runtimeSettingsFile, error)
	update(context.Context, *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateDynamicIPProviders(context.Context, []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateEgressProfiles(context.Context, []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateIngressRules(context.Context, []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateInUserRules(context.Context, []*proxyruntimev1.EgressProfileSettings, []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
}

type runtimeSettingsApplication struct {
	logger                     *slog.Logger
	settings                   runtimeSettingsRepository
	proxyUsers                 []config.ProxyUserRoute
	ipFraudProviderViews       func() []*proxyruntimev1.ProxyIPFraudProviderDescriptor
	ipGeoProviderViews         func() []*proxyruntimev1.ProxyIPGeoProviderDescriptor
	loadMihomoNativeSettings   func(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	updateMihomoNativeSettings func(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	scheduleApply              func([]string)
}

func newRuntimeSettingsApplication(runtime *Runtime) runtimeSettingsApplication {
	return runtimeSettingsApplication{
		logger:               runtime.logger,
		settings:             runtime.settings,
		proxyUsers:           append([]config.ProxyUserRoute(nil), runtime.cfg.ProxyUsers...),
		ipFraudProviderViews: runtime.ipFraudProviders.ProviderDescriptors,
		ipGeoProviderViews:   runtime.ipGeoProviders.ProviderDescriptors,
		loadMihomoNativeSettings: func(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
			return mihomoNativeSettings(ctx, runtime)
		},
		updateMihomoNativeSettings: func(ctx context.Context, config *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
			return updateMihomoNativeSettings(ctx, runtime, config)
		},
		scheduleApply: runtime.scheduleRuntimeSettingsApply,
	}
}

func (a runtimeSettingsApplication) ListProxyIPFraudProviders(context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPFraudProvidersResponse{Providers: a.listIPFraudProviderViews()}, nil
}

func (a runtimeSettingsApplication) ListProxyIPGeoProviders(context.Context) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPGeoProvidersResponse{Providers: a.listIPGeoProviderViews()}, nil
}

func (a runtimeSettingsApplication) GetProxyRuntimeSettings(ctx context.Context) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	settings, err := a.settings.view(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) listIPFraudProviderViews() []*proxyruntimev1.ProxyIPFraudProviderDescriptor {
	if a.ipFraudProviderViews == nil {
		return nil
	}
	return a.ipFraudProviderViews()
}

func (a runtimeSettingsApplication) listIPGeoProviderViews() []*proxyruntimev1.ProxyIPGeoProviderDescriptor {
	if a.ipGeoProviderViews == nil {
		return nil
	}
	return a.ipGeoProviderViews()
}
