package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func enabledDynamicProviderEndpointIDs(settings *runtimeSettingsFile) map[string]map[string]struct{} {
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
			endpointID := dynamic.EndpointIDFromURL(endpoint.GetEndpointUrl())
			if endpointID != "" {
				out[dynamicProviderID][endpointID] = struct{}{}
			}
		}
	}
	return out
}

func enabledEgressProfileIDsFromProfiles(profiles []*proxyruntimev1.EgressProfileSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, profile := range profiles {
		if id := appcore.RuntimeSafeID(profile.GetProfileId()); id != "" && profile.GetEnabled() {
			out[id] = struct{}{}
		}
	}
	return out
}
