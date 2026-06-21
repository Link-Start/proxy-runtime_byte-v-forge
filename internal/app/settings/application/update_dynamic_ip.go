package application

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (a Application) UpdateDynamicIPProviders(ctx context.Context, providers []*proxygatewayv1.ProxyDynamicIPProviderSettings) (*proxygatewayv1.UpdateProxyGatewaySettingsResponse, error) {
	settings, err := a.updateAndSchedule(ctx, func(repository Repository) (*proxygatewayv1.ProxyGatewaySettings, error) {
		return repository.UpdateDynamicIPProviders(ctx, providers)
	})
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.UpdateProxyGatewaySettingsResponse{Settings: settings}, nil
}
