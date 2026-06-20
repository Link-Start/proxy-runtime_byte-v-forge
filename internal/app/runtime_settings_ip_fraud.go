package app

import (
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func ipFraudProviderSecrets(settings *runtimeSettingsFile, providers *ipfraud.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range settingscore.NormalizeRuntimeSettingsWithProviders(settings, providers, nil).GetIpFraudProviders() {
		secrets[ipFraudProviderSecretKey(item.GetKind(), item.GetProviderId())] = settingscore.CleanIPFraudSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func ipFraudProviderSecretKey(kind proxyruntimev1.ProxyIPFraudProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}
