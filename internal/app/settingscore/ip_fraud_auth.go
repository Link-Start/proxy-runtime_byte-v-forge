package settingscore

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

// IPFraudProviders resolves the configured ip-fraud providers from settings,
// loading API-key secrets through resolver.
func IPFraudProviders(ctx context.Context, resolver secretref.Resolver, settings *proxyruntimev1.ProxyRuntimePersistentSettings, registry *ipfraud.Registry) ([]ipfraud.ProviderConfig, error) {
	items := NormalizeRuntimeSettingsWithProviders(settings, registry, nil).GetIpFraudProviders()
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

func ipFraudAuth(ctx context.Context, resolver secretref.Resolver, provider *proxyruntimev1.ProxyIPFraudProviderSettings, registry *ipfraud.Registry) (ipfraud.AuthConfig, error) {
	if provider.GetAnonymous() {
		return ipfraud.AuthConfig{Anonymous: &ipfraud.AnonymousAuthConfig{}}, nil
	}
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return ipfraud.AuthConfig{}, nil
	}
	values, err := appcore.ResolveRuntimeSecretRefs(ctx, resolver, provider.GetApiKeySecretRefs(), IPFraudAPIKeyPurpose)
	if err != nil {
		return ipfraud.AuthConfig{}, err
	}
	return plugin.Auth(values, false), nil
}
