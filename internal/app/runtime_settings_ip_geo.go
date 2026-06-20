package app

import (
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func ipGeoProviderSecrets(settings *runtimeSettingsFile, registry *ipgeo.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range settingscore.NormalizeRuntimeSettingsWithProviders(settings, nil, registry).GetIpGeoProviders() {
		secrets[ipGeoProviderSecretKey(item.GetKind(), item.GetProviderId())] = settingscore.CleanIPGeoSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func ipGeoProviderSecretKey(kind proxyruntimev1.ProxyIPGeoProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}
