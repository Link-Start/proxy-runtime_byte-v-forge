package kernel

import (
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/common/v1"
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

const IPFraudAPIKeyPurpose = "ip_fraud_api_key"

func NormalizeIPFraudProvider(provider *proxygatewayv1.ProxyIPFraudProviderSettings, index int, registry *ipfraud.Registry) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
	provider.ApiKeySecretRefs = CleanIPFraudSecretRefs(provider.GetApiKeySecretRefs())
	if provider.ProviderId == "" {
		provider.ProviderId = registry.DefaultProviderID(provider.GetKind())
	}
	if provider.Weight == 0 {
		provider.Weight = ProviderDefaultWeight(provider.GetKind(), index, registry)
	}
}

func CleanIPFraudSecretRefs(values []*commonv1.SecretRef) []*commonv1.SecretRef {
	return appcore.CleanSecretRefs(values, "proxy-gateway", IPFraudAPIKeyPurpose)
}

func ProviderDefaultWeight(kind proxygatewayv1.ProxyIPFraudProviderKind, index int, registry *ipfraud.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return DefaultProviderWeight(index)
}

func SupportedIPFraudProviders(providers []*proxygatewayv1.ProxyIPFraudProviderSettings, registry *ipfraud.Registry) []*proxygatewayv1.ProxyIPFraudProviderSettings {
	out := make([]*proxygatewayv1.ProxyIPFraudProviderSettings, 0, len(providers))
	for _, provider := range providers {
		if registry.IsProviderKindSupported(provider.GetKind()) {
			out = append(out, provider)
		}
	}
	return out
}
