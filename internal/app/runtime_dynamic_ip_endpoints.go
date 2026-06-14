package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
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
