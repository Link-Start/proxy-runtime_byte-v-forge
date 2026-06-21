package kernel

import (
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/common/v1"
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipgeo"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

const IPGeoAPIKeyPurpose = "ip_geo_api_key"

func NormalizeIPGeoProvider(provider *proxygatewayv1.ProxyIPGeoProviderSettings, index int, registry *ipgeo.Registry) {
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
	return appcore.CleanSecretRefs(values, "proxy-gateway", IPGeoAPIKeyPurpose)
}

func IPGeoProviderDefaultWeight(kind proxygatewayv1.ProxyIPGeoProviderKind, index int, registry *ipgeo.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return DefaultProviderWeight(index)
}

func SupportedIPGeoProviders(providers []*proxygatewayv1.ProxyIPGeoProviderSettings, registry *ipgeo.Registry) []*proxygatewayv1.ProxyIPGeoProviderSettings {
	out := make([]*proxygatewayv1.ProxyIPGeoProviderSettings, 0, len(providers))
	for _, provider := range providers {
		if registry.IsProviderKindSupported(provider.GetKind()) {
			out = append(out, provider)
		}
	}
	return out
}
