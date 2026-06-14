package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) ListProxyIPFraudProviders(context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPFraudProvidersResponse{Providers: a.listIPFraudProviderViews()}, nil
}

func (a runtimeSettingsApplication) ListProxyIPGeoProviders(context.Context) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPGeoProvidersResponse{Providers: a.listIPGeoProviderViews()}, nil
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
