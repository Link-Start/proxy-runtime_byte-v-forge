package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
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
		provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
		provider.ApiKeySecretRefs = cleanIPFraudSecretRefs(provider.GetApiKeySecretRefs())
	}
	for _, provider := range settings.IpGeoProviders {
		if provider == nil {
			continue
		}
		provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
		provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
		provider.ApiKeySecretRefs = cleanIPGeoSecretRefs(provider.GetApiKeySecretRefs())
	}
	for index := range settings.DynamicIpProviders {
		normalizeDynamicIPProvider(settings.DynamicIpProviders[index])
	}
	for index := range settings.EgressProfiles {
		normalizeEgressProfile(settings.EgressProfiles[index])
	}
	for index := range settings.IngressRules {
		settings.IngressRules[index] = ingressRuleFromProto(settings.IngressRules[index], index)
	}
	settings.CheckSettings = normalizeCheckSettings(settings.GetCheckSettings())
	return settings
}

func normalizeRuntimeSettingsWithProviders(settings *runtimeSettingsFile, fraud *ipfraud.Registry, geo *ipgeo.Registry) *runtimeSettingsFile {
	settings = normalizeRuntimeSettings(settings)
	if fraud != nil {
		for index := range settings.IpFraudProviders {
			normalizeIPFraudProvider(settings.IpFraudProviders[index], index, fraud)
		}
		settings.IpFraudProviders = supportedIPFraudProviders(settings.IpFraudProviders, fraud)
	}
	if geo != nil {
		for index := range settings.IpGeoProviders {
			normalizeIPGeoProvider(settings.IpGeoProviders[index], index, geo)
		}
		settings.IpGeoProviders = supportedIPGeoProviders(settings.IpGeoProviders, geo)
	}
	return settings
}
