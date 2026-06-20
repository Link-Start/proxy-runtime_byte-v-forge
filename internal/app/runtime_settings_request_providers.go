package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func dynamicIPProvidersFromRequest(req []*proxyruntimev1.ProxyDynamicIPProviderSettings, registry *providerregistry.Registry) ([]*proxyruntimev1.ProxyDynamicIPProviderSettings, error) {
	seenProviders := map[string]struct{}{}
	out := make([]*proxyruntimev1.ProxyDynamicIPProviderSettings, 0, len(req))
	for index, provider := range req {
		item := dynamicIPProviderFromProto(provider)
		if err := validateDynamicIPProvider(item, index, registry); err != nil {
			return nil, err
		}
		id := kernel.DynamicIPProviderID(item)
		if _, exists := seenProviders[id]; exists {
			return nil, fmt.Errorf("dynamic_ip_providers[%d] duplicates dynamic provider %q", index, id)
		}
		seenProviders[id] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}
