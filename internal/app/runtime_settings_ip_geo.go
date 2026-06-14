package app

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

const ipGeoAPIKeyPurpose = "ip_geo_api_key"

func ipGeoProviderSecrets(settings *runtimeSettingsFile, registry *ipgeo.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range normalizeRuntimeSettingsWithProviders(settings, nil, registry).GetIpGeoProviders() {
		secrets[ipGeoProviderSecretKey(item.GetKind(), item.GetProviderId())] = cleanIPGeoSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func ipGeoProviderSecretKey(kind proxyruntimev1.ProxyIPGeoProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}

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

func normalizeIPGeoProvider(provider *proxyruntimev1.ProxyIPGeoProviderSettings, index int, registry *ipgeo.Registry) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
	provider.ApiKeySecretRefs = cleanIPGeoSecretRefs(provider.GetApiKeySecretRefs())
	if provider.ProviderId == "" {
		provider.ProviderId = registry.DefaultProviderID(provider.GetKind())
	}
	if provider.Weight == 0 {
		provider.Weight = ipGeoProviderDefaultWeight(provider.GetKind(), index, registry)
	}
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

func cleanIPGeoSecretRefs(values []*commonv1.SecretRef) []*commonv1.SecretRef {
	return cleanSecretRefs(values, "proxy-runtime", ipGeoAPIKeyPurpose)
}

func ipGeoProviderDefaultWeight(kind proxyruntimev1.ProxyIPGeoProviderKind, index int, registry *ipgeo.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return defaultProviderWeight(index)
}
