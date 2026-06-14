package app

import (
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
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

func cleanIPGeoSecretRefs(values []*commonv1.SecretRef) []*commonv1.SecretRef {
	return cleanSecretRefs(values, "proxy-runtime", ipGeoAPIKeyPurpose)
}

func ipGeoProviderDefaultWeight(kind proxyruntimev1.ProxyIPGeoProviderKind, index int, registry *ipgeo.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return defaultProviderWeight(index)
}
