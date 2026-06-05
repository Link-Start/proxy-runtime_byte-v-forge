package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

func dynamicIPGatewayMap(settings *runtimeSettingsFile) map[string][]accountproxy.Gateway {
	out := map[string][]accountproxy.Gateway{}
	for _, provider := range normalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		out[provider.GetProviderId()] = accountProxyGateways(provider.GetGateways())
	}
	return out
}

func dynamicIPGateways(settings *runtimeSettingsFile, providerID string) []accountproxy.Gateway {
	return dynamicIPGatewayMap(settings)[strings.TrimSpace(providerID)]
}

func dynamicIPProviderFromProto(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPProviderSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPProviderSettings{
		ProviderId: strings.TrimSpace(in.GetProviderId()),
		Gateways:   make([]*proxyruntimev1.ProxyDynamicIPGatewaySettings, 0, len(in.GetGateways())),
	}
	for _, gateway := range in.GetGateways() {
		out.Gateways = append(out.Gateways, dynamicIPGatewayFromProto(gateway))
	}
	normalizeDynamicIPProvider(out)
	return out
}

func dynamicIPGatewayFromProto(in *proxyruntimev1.ProxyDynamicIPGatewaySettings) *proxyruntimev1.ProxyDynamicIPGatewaySettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPGatewaySettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPGatewaySettings{
		EndpointUrl: normalizeEndpointURL(in.GetEndpointUrl()),
	}
	return out
}

func normalizeDynamicIPProvider(provider *proxyruntimev1.ProxyDynamicIPProviderSettings) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	for index := range provider.Gateways {
		normalizeDynamicIPGateway(provider.Gateways[index], index)
	}
}

func validateDynamicIPProvider(provider *proxyruntimev1.ProxyDynamicIPProviderSettings, index int, accountProviders *providerregistry.Registry) error {
	if !accountProviders.IsSupported(provider.GetProviderId()) {
		return fmt.Errorf("dynamic_ip_providers[%d].provider_id is unsupported", index)
	}
	seen := map[string]struct{}{}
	for gatewayIndex, gateway := range provider.GetGateways() {
		endpointURL := normalizeEndpointURL(gateway.GetEndpointUrl())
		if endpointURL == "" {
			return fmt.Errorf("dynamic_ip_providers[%d].gateways[%d].endpoint_url is required", index, gatewayIndex)
		}
		if _, exists := seen[endpointURL]; exists {
			return fmt.Errorf("dynamic_ip_providers[%d].gateways[%d] duplicates endpoint %q", index, gatewayIndex, endpointURL)
		}
		seen[endpointURL] = struct{}{}
	}
	return nil
}

func normalizeDynamicIPGateway(gateway *proxyruntimev1.ProxyDynamicIPGatewaySettings, _ int) {
	if gateway == nil {
		return
	}
	gateway.EndpointUrl = normalizeEndpointURL(gateway.GetEndpointUrl())
}

func accountProxyGateways(gateways []*proxyruntimev1.ProxyDynamicIPGatewaySettings) []accountproxy.Gateway {
	out := make([]accountproxy.Gateway, 0, len(gateways))
	for _, gateway := range gateways {
		endpointURL := normalizeEndpointURL(gateway.GetEndpointUrl())
		if endpointURL == "" {
			continue
		}
		out = append(out, accountproxy.Gateway{
			ID:          gatewayIDFromEndpointURL(endpointURL),
			EndpointURL: endpointURL,
		})
	}
	return out
}

func cloneDynamicIPProvider(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	return dynamicIPProviderFromProto(in)
}

func normalizeEndpointURL(value string) string {
	return strings.TrimSpace(value)
}

func gatewayIDFromEndpointURL(value string) string {
	value = normalizeEndpointURL(value)
	if value == "" {
		return ""
	}
	return "endpoint-" + shortHash(value)
}
