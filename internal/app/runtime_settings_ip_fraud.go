package app

import (
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

const ipFraudAPIKeyPurpose = "ip_fraud_api_key"

func ipFraudProviderSecrets(settings *runtimeSettingsFile, providers *ipfraud.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range normalizeRuntimeSettingsWithProviders(settings, providers, nil).GetIpFraudProviders() {
		secrets[ipFraudProviderSecretKey(item.GetKind(), item.GetProviderId())] = cleanIPFraudSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func ipFraudProviderSecretKey(kind proxyruntimev1.ProxyIPFraudProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}

func normalizeIPFraudProvider(provider *proxyruntimev1.ProxyIPFraudProviderSettings, index int, registry *ipfraud.Registry) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
	provider.ApiKeySecretRefs = cleanIPFraudSecretRefs(provider.GetApiKeySecretRefs())
	if provider.ProviderId == "" {
		provider.ProviderId = registry.DefaultProviderID(provider.GetKind())
	}
	if provider.Weight == 0 {
		provider.Weight = providerDefaultWeight(provider.GetKind(), index, registry)
	}
}

func cleanIPFraudSecretRefs(values []*commonv1.SecretRef) []*commonv1.SecretRef {
	return appcore.CleanSecretRefs(values, "proxy-runtime", ipFraudAPIKeyPurpose)
}

func providerDefaultWeight(kind proxyruntimev1.ProxyIPFraudProviderKind, index int, registry *ipfraud.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return defaultProviderWeight(index)
}
