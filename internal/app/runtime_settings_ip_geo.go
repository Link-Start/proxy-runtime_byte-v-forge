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

func validateIPGeoProvider(provider *proxyruntimev1.ProxyIPGeoProviderSettings, index int, registry *ipgeo.Registry) error {
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return fmt.Errorf("ip_geo_providers[%d].kind is required", index)
	}
	if provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 {
		return fmt.Errorf("ip_geo_providers[%d] must use anonymous or api key mode, not both", index)
	}
	if provider.GetAnonymous() && !plugin.SupportsAnonymous() {
		return fmt.Errorf("ip_geo_providers[%d] does not support anonymous mode", index)
	}
	if !provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 && !plugin.SupportsAPIKey() {
		return fmt.Errorf("ip_geo_providers[%d] does not support api key mode", index)
	}
	return nil
}

func supportedIPGeoProviders(providers []*proxyruntimev1.ProxyIPGeoProviderSettings, registry *ipgeo.Registry) []*proxyruntimev1.ProxyIPGeoProviderSettings {
	out := make([]*proxyruntimev1.ProxyIPGeoProviderSettings, 0, len(providers))
	for _, provider := range providers {
		if registry.IsProviderKindSupported(provider.GetKind()) {
			out = append(out, provider)
		}
	}
	return out
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
