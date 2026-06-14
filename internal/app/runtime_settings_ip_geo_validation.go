package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
)

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
