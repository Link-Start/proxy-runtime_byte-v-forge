package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func runtimeSettingsView(settings *runtimeSettingsFile) *proxyruntimev1.ProxyRuntimeSettings {
	settings = normalizeRuntimeSettings(settings)
	edge := settings.GetEdgeCanary()
	out := &proxyruntimev1.ProxyRuntimeSettings{
		EdgeCanary: &proxyruntimev1.ProxyEdgeCanarySettingsView{
			Url:             edge.GetUrl(),
			TokenConfigured: secretRefConfigured(edge.GetTokenSecretRef()),
			Enabled:         edgeCanaryEnabled(edge),
		},
		CheckSettings: cloneCheckSettings(settings.GetCheckSettings()),
	}
	for _, provider := range settings.GetIpFraudProviders() {
		out.IpFraudProviders = append(out.IpFraudProviders, &proxyruntimev1.ProxyIPFraudProviderSettingsView{
			ProviderId:       provider.GetProviderId(),
			Weight:           provider.GetWeight(),
			Kind:             provider.GetKind(),
			DisplayName:      provider.GetDisplayName(),
			Anonymous:        provider.GetAnonymous(),
			ApiKeyConfigured: len(provider.GetApiKeySecretRefs()) > 0,
			ApiKeyCount:      uint32(len(provider.GetApiKeySecretRefs())),
		})
	}
	for _, provider := range settings.GetIpGeoProviders() {
		out.IpGeoProviders = append(out.IpGeoProviders, &proxyruntimev1.ProxyIPGeoProviderSettingsView{
			ProviderId:       provider.GetProviderId(),
			Weight:           provider.GetWeight(),
			Kind:             provider.GetKind(),
			DisplayName:      provider.GetDisplayName(),
			Anonymous:        provider.GetAnonymous(),
			ApiKeyConfigured: len(provider.GetApiKeySecretRefs()) > 0,
			ApiKeyCount:      uint32(len(provider.GetApiKeySecretRefs())),
		})
	}
	for _, provider := range settings.GetDynamicIpProviders() {
		out.DynamicIpProviders = append(out.DynamicIpProviders, cloneDynamicIPProvider(provider))
	}
	for _, profile := range settings.GetEgressProfiles() {
		out.EgressProfiles = append(out.EgressProfiles, cloneEgressProfile(profile))
	}
	for _, rule := range settings.GetIngressRules() {
		out.IngressRules = append(out.IngressRules, cloneIngressRule(rule))
	}
	return out
}
