package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsApplication struct {
	runtime *Runtime
}

func newRuntimeSettingsApplication(runtime *Runtime) runtimeSettingsApplication {
	return runtimeSettingsApplication{runtime: runtime}
}

func (a runtimeSettingsApplication) ListProxyIPFraudProviders(context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPFraudProvidersResponse{Providers: a.runtime.ipFraudProviders.ProviderDescriptors()}, nil
}

func (a runtimeSettingsApplication) ListProxyIPGeoProviders(context.Context) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPGeoProvidersResponse{Providers: a.runtime.ipGeoProviders.ProviderDescriptors()}, nil
}

func (a runtimeSettingsApplication) GetProxyRuntimeSettings(ctx context.Context) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	settings, err := a.runtime.settings.view(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyRuntimeSettingsResponse{Settings: settings}, nil
}
