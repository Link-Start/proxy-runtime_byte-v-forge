package domain

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func enabledEgressProfileIDsFromProfiles(profiles []*proxygatewayv1.EgressProfileSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, profile := range profiles {
		if id := appcore.RuntimeSafeID(profile.GetProfileId()); id != "" && profile.GetEnabled() {
			out[id] = struct{}{}
		}
	}
	return out
}

func EnabledDynamicProviderEndpointIDs(settings *proxygatewayv1.ProxyGatewayPersistentSettings) map[string]map[string]struct{} {
	out := map[string]map[string]struct{}{}
	for _, provider := range kernel.NormalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		dynamicProviderID := kernel.DynamicIPProviderID(provider)
		if dynamicProviderID == "" {
			continue
		}
		if out[dynamicProviderID] == nil {
			out[dynamicProviderID] = map[string]struct{}{}
		}
		for _, endpoint := range provider.GetEndpoints() {
			endpointID := kernel.EndpointIDFromURL(endpoint.GetEndpointUrl())
			if endpointID != "" {
				out[dynamicProviderID][endpointID] = struct{}{}
			}
		}
	}
	return out
}
