package settingscore

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

// IPGeoProvidersFromRequest builds the persisted ip-geo provider settings from
// an update request, writing inline API keys as secrets and rejecting duplicate
// providers.
func IPGeoProvidersFromRequest(ctx context.Context, writer secretref.Writer, req []*proxyruntimev1.ProxyIPGeoProviderSettings, current *proxyruntimev1.ProxyRuntimePersistentSettings, registry *ipgeo.Registry) ([]*proxyruntimev1.ProxyIPGeoProviderSettings, error) {
	currentProviders := ipGeoProviderSecrets(current, registry)
	seenProviders := map[string]struct{}{}
	out := make([]*proxyruntimev1.ProxyIPGeoProviderSettings, 0, len(req))
	for index, provider := range req {
		item, err := ipGeoProviderFromRequest(ctx, writer, provider, currentProviders, index, registry)
		if err != nil {
			return nil, err
		}
		if err := validateIPGeoProvider(item, index, registry); err != nil {
			return nil, err
		}
		key := ipGeoProviderSecretKey(item.GetKind(), item.GetProviderId())
		if _, exists := seenProviders[key]; exists {
			return nil, fmt.Errorf("ip_geo_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seenProviders[key] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}

func ipGeoProviderSecrets(settings *proxyruntimev1.ProxyRuntimePersistentSettings, registry *ipgeo.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range NormalizeRuntimeSettingsWithProviders(settings, nil, registry).GetIpGeoProviders() {
		secrets[ipGeoProviderSecretKey(item.GetKind(), item.GetProviderId())] = CleanIPGeoSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func ipGeoProviderSecretKey(kind proxyruntimev1.ProxyIPGeoProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}

func ipGeoProviderFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPGeoProviderSettings, current map[string][]*commonv1.SecretRef, index int, registry *ipgeo.Registry) (*proxyruntimev1.ProxyIPGeoProviderSettings, error) {
	if in == nil {
		return &proxyruntimev1.ProxyIPGeoProviderSettings{}, nil
	}
	id := strings.TrimSpace(in.GetProviderId())
	if id == "" {
		id = registry.DefaultProviderID(in.GetKind())
	}
	apiKeySecretRefs, err := ipGeoSecretRefsFromRequest(ctx, writer, in, id)
	if err != nil {
		return nil, err
	}
	if len(apiKeySecretRefs) == 0 && !in.GetClearApiKeys() && !in.GetAnonymous() {
		apiKeySecretRefs = current[ipGeoProviderSecretKey(in.GetKind(), id)]
	}
	weight := in.GetWeight()
	if weight == 0 {
		weight = IPGeoProviderDefaultWeight(in.GetKind(), index, registry)
	}
	return &proxyruntimev1.ProxyIPGeoProviderSettings{
		ProviderId:       id,
		DisplayName:      strings.TrimSpace(in.GetDisplayName()),
		Weight:           weight,
		Kind:             in.GetKind(),
		Anonymous:        in.GetAnonymous(),
		ApiKeySecretRefs: apiKeySecretRefs,
	}, nil
}

func ipGeoSecretRefsFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPGeoProviderSettings, providerID string) ([]*commonv1.SecretRef, error) {
	if refs := CleanIPGeoSecretRefs(in.GetApiKeySecretRefs()); len(refs) > 0 {
		return refs, nil
	}
	rawValues := appcore.CleanList(in.GetApiKeyValues())
	if len(rawValues) == 0 {
		return nil, nil
	}
	out := make([]*commonv1.SecretRef, 0, len(rawValues))
	for index, raw := range rawValues {
		secretID := secretref.StableID("proxy-runtime-ip-geo-api-key", fmt.Sprintf("%d", in.GetKind()), providerID, fmt.Sprintf("%d", index))
		saved, err := WriteRuntimeSecret(ctx, writer, raw, secretID, IPGeoAPIKeyPurpose)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}
	return CleanIPGeoSecretRefs(out), nil
}

func validateIPGeoProvider(provider *proxyruntimev1.ProxyIPGeoProviderSettings, index int, registry *ipgeo.Registry) error {
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return fmt.Errorf("ip_geo_providers[%d].kind is required", index)
	}
	if provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 {
		return fmt.Errorf("ip_geo_providers[%d] must use anonymous or api key mode, not both", index)
	}
	if provider.GetAnonymous() && !plugin.SupportsAnonymous() {
		return fmt.Errorf("ip_geo_providers[%d] does not support anonymous mode", index)
	}
	if !provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 && !plugin.SupportsAPIKey() {
		return fmt.Errorf("ip_geo_providers[%d] does not support api key mode", index)
	}
	return nil
}
