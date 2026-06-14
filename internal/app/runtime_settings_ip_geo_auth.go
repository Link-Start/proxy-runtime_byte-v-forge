package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func ipGeoProviders(ctx context.Context, resolver secretref.Resolver, settings *runtimeSettingsFile, registry *ipgeo.Registry) ([]ipgeo.ProviderConfig, error) {
	items := normalizeRuntimeSettingsWithProviders(settings, nil, registry).GetIpGeoProviders()
	providers := make([]ipgeo.ProviderConfig, 0, len(items))
	for _, item := range items {
		if !item.GetAnonymous() && len(item.GetApiKeySecretRefs()) == 0 {
			continue
		}
		auth, err := ipGeoAuth(ctx, resolver, item, registry)
		if err != nil {
			return nil, err
		}
		providers = append(providers, ipgeo.ProviderConfig{
			ID:     item.GetProviderId(),
			Kind:   item.GetKind(),
			Weight: int(item.GetWeight()),
			Auth:   auth,
		})
	}
	return providers, nil
}

func ipGeoAuth(ctx context.Context, resolver secretref.Resolver, provider *proxyruntimev1.ProxyIPGeoProviderSettings, registry *ipgeo.Registry) (ipgeo.AuthConfig, error) {
	if provider.GetAnonymous() {
		return ipgeo.AuthConfig{Anonymous: &ipgeo.AnonymousAuthConfig{}}, nil
	}
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return ipgeo.AuthConfig{}, nil
	}
	values, err := resolveRuntimeSecretRefs(ctx, resolver, provider.GetApiKeySecretRefs(), ipGeoAPIKeyPurpose)
	if err != nil {
		return ipgeo.AuthConfig{}, err
	}
	return plugin.Auth(values, false), nil
}
