package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

func dynamicIPEndpoints(settings *runtimeSettingsFile, dynamicProviderID string, providerID string) []accountproxy.Gateway {
	dynamicProviderID = runtimeSafeID(dynamicProviderID)
	providerID = strings.TrimSpace(providerID)
	out := []accountproxy.Gateway{}
	seen := map[string]struct{}{}
	for _, provider := range dynamicIPProviderInstances(settings) {
		if dynamicProviderID != "" && provider.dynamicProviderID != dynamicProviderID {
			continue
		}
		if providerID != "" && provider.providerID != providerID {
			continue
		}
		for _, endpoint := range provider.endpoints {
			key := firstNonEmpty(endpoint.ID, endpointIDFromURL(endpoint.EndpointURL), endpoint.EndpointURL)
			if key == "" {
				continue
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, endpoint)
		}
	}
	return out
}

func dynamicIPEndpointMap(settings *runtimeSettingsFile) map[string][]accountproxy.Gateway {
	out := map[string][]accountproxy.Gateway{}
	seen := map[string]map[string]struct{}{}
	for _, provider := range dynamicIPProviderInstances(settings) {
		if provider.providerID == "" {
			continue
		}
		if seen[provider.providerID] == nil {
			seen[provider.providerID] = map[string]struct{}{}
		}
		for _, endpoint := range provider.endpoints {
			key := firstNonEmpty(endpoint.ID, endpointIDFromURL(endpoint.EndpointURL), endpoint.EndpointURL)
			if key == "" {
				continue
			}
			if _, exists := seen[provider.providerID][key]; exists {
				continue
			}
			seen[provider.providerID][key] = struct{}{}
			out[provider.providerID] = append(out[provider.providerID], endpoint)
		}
	}
	return out
}

type dynamicIPProviderInstance struct {
	dynamicProviderID        string
	providerID               string
	displayName              string
	rotatingConcurrencyLimit uint32
	stickyConcurrencyLimit   uint32
	endpoints                []accountproxy.Gateway
}

func dynamicIPProviderInstances(settings *runtimeSettingsFile) []dynamicIPProviderInstance {
	out := []dynamicIPProviderInstance{}
	for _, provider := range normalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		out = append(out, dynamicIPProviderInstance{
			dynamicProviderID:        dynamicIPProviderID(provider),
			providerID:               strings.TrimSpace(provider.GetProviderId()),
			displayName:              strings.TrimSpace(provider.GetDisplayName()),
			rotatingConcurrencyLimit: normalizeDynamicProviderRotatingConcurrencyLimit(provider.GetRotatingConcurrencyLimit()),
			stickyConcurrencyLimit:   normalizeDynamicProviderStickyConcurrencyLimit(provider.GetStickyConcurrencyLimit()),
			endpoints:                accountProxyEndpoints(provider.GetEndpoints()),
		})
	}
	return out
}

func dynamicIPProviderFromProto(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPProviderSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPProviderSettings{
		ProviderId:               strings.TrimSpace(in.GetProviderId()),
		DynamicProviderId:        runtimeSafeID(in.GetDynamicProviderId()),
		DisplayName:              strings.TrimSpace(in.GetDisplayName()),
		RotatingConcurrencyLimit: normalizeDynamicProviderRotatingConcurrencyLimit(in.GetRotatingConcurrencyLimit()),
		StickyConcurrencyLimit:   normalizeDynamicProviderStickyConcurrencyLimit(in.GetStickyConcurrencyLimit()),
		Endpoints:                make([]*proxyruntimev1.ProxyDynamicIPEndpointSettings, 0, len(in.GetEndpoints())),
	}
	for _, endpoint := range in.GetEndpoints() {
		out.Endpoints = append(out.Endpoints, dynamicIPEndpointFromProto(endpoint))
	}
	normalizeDynamicIPProvider(out)
	return out
}

func dynamicIPEndpointFromProto(in *proxyruntimev1.ProxyDynamicIPEndpointSettings) *proxyruntimev1.ProxyDynamicIPEndpointSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPEndpointSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPEndpointSettings{
		EndpointUrl: normalizeEndpointURL(in.GetEndpointUrl()),
	}
	return out
}

func normalizeDynamicIPProvider(provider *proxyruntimev1.ProxyDynamicIPProviderSettings) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.DynamicProviderId = runtimeSafeID(provider.GetDynamicProviderId())
	if provider.DynamicProviderId == "" {
		provider.DynamicProviderId = runtimeSafeID(provider.GetProviderId())
	}
	provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
	if provider.DisplayName == "" {
		provider.DisplayName = provider.GetDynamicProviderId()
	}
	provider.RotatingConcurrencyLimit = normalizeDynamicProviderRotatingConcurrencyLimit(provider.GetRotatingConcurrencyLimit())
	provider.StickyConcurrencyLimit = normalizeDynamicProviderStickyConcurrencyLimit(provider.GetStickyConcurrencyLimit())
	for index := range provider.Endpoints {
		normalizeDynamicIPEndpoint(provider.Endpoints[index])
	}
}

func validateDynamicIPProvider(provider *proxyruntimev1.ProxyDynamicIPProviderSettings, index int, accountProviders *providerregistry.Registry) error {
	if dynamicIPProviderID(provider) == "" {
		return fmt.Errorf("dynamic_ip_providers[%d].dynamic_provider_id is required", index)
	}
	if !accountProviders.IsSupported(provider.GetProviderId()) {
		return fmt.Errorf("dynamic_ip_providers[%d].provider_id is unsupported", index)
	}
	seen := map[string]struct{}{}
	for endpointIndex, endpoint := range provider.GetEndpoints() {
		endpointURL := normalizeEndpointURL(endpoint.GetEndpointUrl())
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

func normalizeDynamicIPEndpoint(endpoint *proxyruntimev1.ProxyDynamicIPEndpointSettings) {
	if endpoint == nil {
		return
	}
	endpoint.EndpointUrl = normalizeEndpointURL(endpoint.GetEndpointUrl())
}

func accountProxyEndpoints(endpoints []*proxyruntimev1.ProxyDynamicIPEndpointSettings) []accountproxy.Gateway {
	out := make([]accountproxy.Gateway, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointURL := normalizeEndpointURL(endpoint.GetEndpointUrl())
		if endpointURL == "" {
			continue
		}
		out = append(out, accountproxy.Gateway{
			ID:          endpointIDFromURL(endpointURL),
			EndpointURL: endpointURL,
		})
	}
	return out
}

func cloneDynamicIPProvider(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	return dynamicIPProviderFromProto(in)
}

func cloneDynamicIPEndpoints(in []*proxyruntimev1.ProxyDynamicIPEndpointSettings) []*proxyruntimev1.ProxyDynamicIPEndpointSettings {
	out := make([]*proxyruntimev1.ProxyDynamicIPEndpointSettings, 0, len(in))
	for _, endpoint := range in {
		out = append(out, dynamicIPEndpointFromProto(endpoint))
	}
	return out
}

func dynamicIPProviderID(provider *proxyruntimev1.ProxyDynamicIPProviderSettings) string {
	return runtimeSafeID(provider.GetDynamicProviderId())
}

func normalizeEndpointURL(value string) string {
	return strings.TrimSpace(value)
}

func endpointIDFromURL(value string) string {
	value = normalizeEndpointURL(value)
	if value == "" {
		return ""
	}
	return "endpoint-" + shortHash(value)
}
