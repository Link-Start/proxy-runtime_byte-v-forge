package kernel

import (
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

const IPGeoAPIKeyPurpose = "ip_geo_api_key"

func NormalizeIPGeoProvider(provider *proxyruntimev1.ProxyIPGeoProviderSettings, index int, registry *ipgeo.Registry) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
	provider.ApiKeySecretRefs = CleanIPGeoSecretRefs(provider.GetApiKeySecretRefs())
	if provider.ProviderId == "" {
		provider.ProviderId = registry.DefaultProviderID(provider.GetKind())
	}
	if provider.Weight == 0 {
		provider.Weight = IPGeoProviderDefaultWeight(provider.GetKind(), index, registry)
	}
}

func CleanIPGeoSecretRefs(values []*commonv1.SecretRef) []*commonv1.SecretRef {
	return appcore.CleanSecretRefs(values, "proxy-runtime", IPGeoAPIKeyPurpose)
}

func IPGeoProviderDefaultWeight(kind proxyruntimev1.ProxyIPGeoProviderKind, index int, registry *ipgeo.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return DefaultProviderWeight(index)
}

func SupportedIPGeoProviders(providers []*proxyruntimev1.ProxyIPGeoProviderSettings, registry *ipgeo.Registry) []*proxyruntimev1.ProxyIPGeoProviderSettings {
	out := make([]*proxyruntimev1.ProxyIPGeoProviderSettings, 0, len(providers))
	for _, provider := range providers {
		if registry.IsProviderKindSupported(provider.GetKind()) {
			out = append(out, provider)
		}
	}
	return out
}
