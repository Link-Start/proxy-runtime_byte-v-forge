package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
)

func normalizeRuntimeSettings(settings *runtimeSettingsFile) *runtimeSettingsFile {
	if settings == nil {
		settings = &proxyruntimev1.ProxyRuntimePersistentSettings{}
	}
	if settings.EdgeCanary != nil {
		settings.EdgeCanary.Url = strings.TrimSpace(settings.EdgeCanary.GetUrl())
		settings.EdgeCanary.TokenSecretRef = cloneSecretRef(settings.EdgeCanary.GetTokenSecretRef(), "proxy-runtime", "edge_canary_token")
	}
	for _, provider := range settings.IpFraudProviders {
		if provider == nil {
			continue
		}
		provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
		provider.ApiKeySecretRefs = cleanIPFraudSecretRefs(provider.GetApiKeySecretRefs())
	}
	for index := range settings.DynamicIpProviders {
		normalizeDynamicIPProvider(settings.DynamicIpProviders[index])
	}
	settings.CheckSettings = normalizeCheckSettings(settings.GetCheckSettings())
	return settings
}

func normalizeRuntimeSettingsWithProviders(settings *runtimeSettingsFile, ipFraudProviders *ipfraud.Registry) *runtimeSettingsFile {
	settings = normalizeRuntimeSettings(settings)
	for index := range settings.IpFraudProviders {
		normalizeIPFraudProvider(settings.IpFraudProviders[index], index, ipFraudProviders)
	}
	settings.IpFraudProviders = supportedIPFraudProviders(settings.IpFraudProviders, ipFraudProviders)
	return settings
}
