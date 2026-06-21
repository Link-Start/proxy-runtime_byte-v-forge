package kernel

import (
	"fmt"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"
	"github.com/byte-v-forge/proxy-gateway/internal/ipgeo"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func DecodeRuntimeSettings(raw string) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
	settings := &proxygatewayv1.ProxyGatewayPersistentSettings{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode runtime settings: %w", err)
		}
	}
	return NormalizeRuntimeSettings(settings), nil
}

func NormalizeRuntimeSettings(settings *proxygatewayv1.ProxyGatewayPersistentSettings) *proxygatewayv1.ProxyGatewayPersistentSettings {
	if settings == nil {
		settings = &proxygatewayv1.ProxyGatewayPersistentSettings{}
	}
	if settings.EdgeCanary != nil {
		settings.EdgeCanary.Url = strings.TrimSpace(settings.EdgeCanary.GetUrl())
		settings.EdgeCanary.TokenSecretRef = appcore.CloneSecretRef(settings.EdgeCanary.GetTokenSecretRef(), "proxy-gateway", "edge_canary_token")
	}
	for _, provider := range settings.IpFraudProviders {
		if provider == nil {
			continue
		}
		provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
		provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
		provider.ApiKeySecretRefs = CleanIPFraudSecretRefs(provider.GetApiKeySecretRefs())
	}
	for _, provider := range settings.IpGeoProviders {
		if provider == nil {
			continue
		}
		provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
		provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
		provider.ApiKeySecretRefs = CleanIPGeoSecretRefs(provider.GetApiKeySecretRefs())
	}
	for index := range settings.DynamicIpProviders {
		NormalizeDynamicIPProvider(settings.DynamicIpProviders[index])
	}
	for index := range settings.EgressProfiles {
		NormalizeEgressProfile(settings.EgressProfiles[index])
	}
	for index := range settings.IngressRules {
		settings.IngressRules[index] = IngressRuleFromProto(settings.IngressRules[index], index)
	}
	settings.CheckSettings = NormalizeCheckSettings(settings.GetCheckSettings())
	return settings
}

func NormalizeRuntimeSettingsWithProviders(settings *proxygatewayv1.ProxyGatewayPersistentSettings, fraud *ipfraud.Registry, geo *ipgeo.Registry) *proxygatewayv1.ProxyGatewayPersistentSettings {
	settings = NormalizeRuntimeSettings(settings)
	if fraud != nil {
		for index := range settings.IpFraudProviders {
			NormalizeIPFraudProvider(settings.IpFraudProviders[index], index, fraud)
		}
		settings.IpFraudProviders = SupportedIPFraudProviders(settings.IpFraudProviders, fraud)
	}
	if geo != nil {
		for index := range settings.IpGeoProviders {
			NormalizeIPGeoProvider(settings.IpGeoProviders[index], index, geo)
		}
		settings.IpGeoProviders = SupportedIPGeoProviders(settings.IpGeoProviders, geo)
	}
	return settings
}
