package persistence

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	settingssecret "github.com/byte-v-forge/proxy-gateway/internal/app/settings/adapter/secret"
	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

func (s *Store) Update(ctx context.Context, req *proxygatewayv1.UpdateProxyGatewaySettingsRequest) (*proxygatewayv1.ProxyGatewaySettings, error) {
	return s.mutateRuntimeSettings(ctx, func(current *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
		nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
		if err != nil {
			return nil, err
		}
		return settingssecret.SettingsFromRequest(ctx, s.secretWriter, req, current, s.accountProviders, s.ipFraudProviders, s.ipGeoProviders, nativeResourceIDs)
	})
}

func (s *Store) UpdateDynamicIPProviders(ctx context.Context, providers []*proxygatewayv1.ProxyDynamicIPProviderSettings) (*proxygatewayv1.ProxyGatewaySettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
		nextProviders, err := settingsdomain.DynamicIPProvidersFromRequest(providers, s.accountProviders)
		if err != nil {
			return nil, err
		}
		settings.DynamicIpProviders = nextProviders
		return settings, nil
	})
}
