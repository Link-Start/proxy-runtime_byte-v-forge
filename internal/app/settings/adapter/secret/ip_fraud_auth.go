package secret

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

// IPFraudProviders resolves the configured ip-fraud providers from settings,
// loading API-key secrets through resolver.
func IPFraudProviders(ctx context.Context, resolver secretref.Resolver, settings *proxygatewayv1.ProxyGatewayPersistentSettings, registry *ipfraud.Registry) ([]ipfraud.ProviderConfig, error) {
	items := kernel.NormalizeRuntimeSettingsWithProviders(settings, registry, nil).GetIpFraudProviders()
	providers := make([]ipfraud.ProviderConfig, 0, len(items))
	for _, item := range items {
		if !item.GetAnonymous() && len(item.GetApiKeySecretRefs()) == 0 {
			continue
		}
		auth, err := ipFraudAuth(ctx, resolver, item, registry)
		if err != nil {
			return nil, err
		}
		providers = append(providers, ipfraud.ProviderConfig{
			ID:     item.GetProviderId(),
			Kind:   item.GetKind(),
			Weight: int(item.GetWeight()),
			Auth:   auth,
		})
	}
	return providers, nil
}

func ipFraudAuth(ctx context.Context, resolver secretref.Resolver, provider *proxygatewayv1.ProxyIPFraudProviderSettings, registry *ipfraud.Registry) (ipfraud.AuthConfig, error) {
	if provider.GetAnonymous() {
		return ipfraud.AuthConfig{Anonymous: &ipfraud.AnonymousAuthConfig{}}, nil
	}
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return ipfraud.AuthConfig{}, nil
	}
	values, err := appcore.ResolveRuntimeSecretRefs(ctx, resolver, provider.GetApiKeySecretRefs(), kernel.IPFraudAPIKeyPurpose)
	if err != nil {
		return ipfraud.AuthConfig{}, err
	}
	return plugin.Auth(values, false), nil
}
