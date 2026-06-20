package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func validateDynamicIPProvider(provider *proxyruntimev1.ProxyDynamicIPProviderSettings, index int, accountProviders *providerregistry.Registry) error {
	if settingscore.DynamicIPProviderID(provider) == "" {
		return fmt.Errorf("dynamic_ip_providers[%d].dynamic_provider_id is required", index)
	}
	if accountProviders == nil || !accountProviders.IsSupported(provider.GetProviderId()) {
		return fmt.Errorf("dynamic_ip_providers[%d].provider_id is unsupported", index)
	}
	seen := map[string]struct{}{}
	for endpointIndex, endpoint := range provider.GetEndpoints() {
		endpointURL := settingscore.NormalizeEndpointURL(endpoint.GetEndpointUrl())
		if endpointURL == "" {
			return fmt.Errorf("dynamic_ip_providers[%d].endpoints[%d].endpoint_url is required", index, endpointIndex)
		}
		if _, exists := seen[endpointURL]; exists {
			return fmt.Errorf("dynamic_ip_providers[%d].endpoints[%d] duplicates endpoint %q", index, endpointIndex, endpointURL)
		}
		seen[endpointURL] = struct{}{}
	}
	return nil
}
