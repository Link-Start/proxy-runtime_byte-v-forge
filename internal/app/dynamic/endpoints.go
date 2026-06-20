package dynamic

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func Endpoints(settings *proxyruntimev1.ProxyRuntimePersistentSettings, dynamicProviderID string, providerID string) []accountproxy.Gateway {
	dynamicProviderID = appcore.RuntimeSafeID(dynamicProviderID)
	providerID = strings.TrimSpace(providerID)
	out := []accountproxy.Gateway{}
	seen := map[string]struct{}{}
	for _, provider := range ProviderInstances(settings) {
		if dynamicProviderID != "" && provider.DynamicProviderID != dynamicProviderID {
			continue
		}
		if providerID != "" && provider.ProviderID != providerID {
			continue
		}
		for _, endpoint := range provider.Endpoints {
			key := appcore.FirstNonEmpty(endpoint.ID, kernel.EndpointIDFromURL(endpoint.EndpointURL), endpoint.EndpointURL)
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

func EndpointMap(settings *proxyruntimev1.ProxyRuntimePersistentSettings) map[string][]accountproxy.Gateway {
	out := map[string][]accountproxy.Gateway{}
	seen := map[string]map[string]struct{}{}
	for _, provider := range ProviderInstances(settings) {
		if provider.ProviderID == "" {
			continue
		}
		if seen[provider.ProviderID] == nil {
			seen[provider.ProviderID] = map[string]struct{}{}
		}
		for _, endpoint := range provider.Endpoints {
			key := appcore.FirstNonEmpty(endpoint.ID, kernel.EndpointIDFromURL(endpoint.EndpointURL), endpoint.EndpointURL)
			if key == "" {
				continue
			}
			if _, exists := seen[provider.ProviderID][key]; exists {
				continue
			}
			seen[provider.ProviderID][key] = struct{}{}
			out[provider.ProviderID] = append(out[provider.ProviderID], endpoint)
		}
	}
	return out
}

type ProviderInstance struct {
	DynamicProviderID        string
	ProviderID               string
	DisplayName              string
	RotatingConcurrencyLimit uint32
	StickyConcurrencyLimit   uint32
	Endpoints                []accountproxy.Gateway
}

func ProviderInstances(settings *proxyruntimev1.ProxyRuntimePersistentSettings) []ProviderInstance {
	out := []ProviderInstance{}
	for _, provider := range kernel.NormalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		out = append(out, ProviderInstance{
			DynamicProviderID:        kernel.DynamicIPProviderID(provider),
			ProviderID:               strings.TrimSpace(provider.GetProviderId()),
			DisplayName:              strings.TrimSpace(provider.GetDisplayName()),
			RotatingConcurrencyLimit: kernel.NormalizeDynamicProviderRotatingConcurrencyLimit(provider.GetRotatingConcurrencyLimit()),
			StickyConcurrencyLimit:   kernel.NormalizeDynamicProviderStickyConcurrencyLimit(provider.GetStickyConcurrencyLimit()),
			Endpoints:                accountProxyEndpoints(provider.GetEndpoints()),
		})
	}
	return out
}

func accountProxyEndpoints(endpoints []*proxyruntimev1.ProxyDynamicIPEndpointSettings) []accountproxy.Gateway {
	out := make([]accountproxy.Gateway, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointURL := kernel.NormalizeEndpointURL(endpoint.GetEndpointUrl())
		if endpointURL == "" {
			continue
		}
		out = append(out, accountproxy.Gateway{
			ID:          kernel.EndpointIDFromURL(endpointURL),
			EndpointURL: endpointURL,
		})
	}
	return out
}

func ProviderInstanceConcurrencyLimit(provider ProviderInstance, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	if leaseapp.ConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return kernel.NormalizeDynamicProviderRotatingConcurrencyLimit(provider.RotatingConcurrencyLimit)
	}
	return kernel.NormalizeDynamicProviderStickyConcurrencyLimit(provider.StickyConcurrencyLimit)
}
