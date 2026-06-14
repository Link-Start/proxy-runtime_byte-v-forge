package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) ListProxyIPFraudProviders(ctx context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return a.settingsUsecase().ListIPFraudProviders(ctx)
}

func (a runtimeSettingsApplication) ListProxyIPGeoProviders(ctx context.Context) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return a.settingsUsecase().ListIPGeoProviders(ctx)
}
