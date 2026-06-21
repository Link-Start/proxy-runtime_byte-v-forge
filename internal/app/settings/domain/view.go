package domain

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func RuntimeSettingsView(settings *proxygatewayv1.ProxyGatewayPersistentSettings) *proxygatewayv1.ProxyGatewaySettings {
	settings = kernel.NormalizeRuntimeSettings(settings)
	edge := settings.GetEdgeCanary()
	out := &proxygatewayv1.ProxyGatewaySettings{
		EdgeCanary: &proxygatewayv1.ProxyEdgeCanarySettingsView{
			Url:             edge.GetUrl(),
			TokenConfigured: appcore.SecretRefConfigured(edge.GetTokenSecretRef()),
			Enabled:         EdgeCanaryEnabled(edge),
		},
		CheckSettings: kernel.CloneCheckSettings(settings.GetCheckSettings()),
	}
	for _, provider := range settings.GetIpFraudProviders() {
		out.IpFraudProviders = append(out.IpFraudProviders, &proxygatewayv1.ProxyIPFraudProviderSettingsView{
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
		out.IpGeoProviders = append(out.IpGeoProviders, &proxygatewayv1.ProxyIPGeoProviderSettingsView{
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
