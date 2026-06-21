package domain

import (
	"fmt"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	providerregistry "github.com/byte-v-forge/proxy-gateway/internal/provider/registry"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func DynamicIPProvidersFromRequest(req []*proxygatewayv1.ProxyDynamicIPProviderSettings, registry *providerregistry.Registry) ([]*proxygatewayv1.ProxyDynamicIPProviderSettings, error) {
	seenProviders := map[string]struct{}{}
	out := make([]*proxygatewayv1.ProxyDynamicIPProviderSettings, 0, len(req))
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
