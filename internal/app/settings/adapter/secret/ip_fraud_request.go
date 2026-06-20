package secret

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

// IPFraudProvidersFromRequest builds the persisted ip-fraud provider settings
// from an update request, writing inline API keys as secrets and rejecting
// duplicate providers.
func IPFraudProvidersFromRequest(ctx context.Context, writer secretref.Writer, req []*proxyruntimev1.ProxyIPFraudProviderSettings, current *proxyruntimev1.ProxyRuntimePersistentSettings, registry *ipfraud.Registry) ([]*proxyruntimev1.ProxyIPFraudProviderSettings, error) {
	currentProviders := ipFraudProviderSecrets(current, registry)
	seenProviders := map[string]struct{}{}
	out := make([]*proxyruntimev1.ProxyIPFraudProviderSettings, 0, len(req))
	for index, provider := range req {
		item, err := ipFraudProviderFromRequest(ctx, writer, provider, currentProviders, index, registry)
		if err != nil {
			return nil, err
		}
		if err := validateIPFraudProvider(item, index, registry); err != nil {
			return nil, err
		}
		key := ipFraudProviderSecretKey(item.GetKind(), item.GetProviderId())
		if _, exists := seenProviders[key]; exists {
			return nil, fmt.Errorf("ip_fraud_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seenProviders[key] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}

func ipFraudProviderSecrets(settings *proxyruntimev1.ProxyRuntimePersistentSettings, providers *ipfraud.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range kernel.NormalizeRuntimeSettingsWithProviders(settings, providers, nil).GetIpFraudProviders() {
		secrets[ipFraudProviderSecretKey(item.GetKind(), item.GetProviderId())] = kernel.CleanIPFraudSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func ipFraudProviderSecretKey(kind proxyruntimev1.ProxyIPFraudProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}

func ipFraudProviderFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPFraudProviderSettings, current map[string][]*commonv1.SecretRef, index int, registry *ipfraud.Registry) (*proxyruntimev1.ProxyIPFraudProviderSettings, error) {
	if in == nil {
		return &proxyruntimev1.ProxyIPFraudProviderSettings{}, nil
	}
	id := strings.TrimSpace(in.GetProviderId())
	if id == "" {
		id = registry.DefaultProviderID(in.GetKind())
	}
	apiKeySecretRefs, err := ipFraudSecretRefsFromRequest(ctx, writer, in, id)
	if err != nil {
		return nil, err
	}
	if len(apiKeySecretRefs) == 0 && !in.GetClearApiKeys() && !in.GetAnonymous() {
		apiKeySecretRefs = current[ipFraudProviderSecretKey(in.GetKind(), id)]
	}
	weight := in.GetWeight()
	if weight == 0 {
		weight = kernel.ProviderDefaultWeight(in.GetKind(), index, registry)
	}
	return &proxyruntimev1.ProxyIPFraudProviderSettings{
		ProviderId:       id,
		DisplayName:      strings.TrimSpace(in.GetDisplayName()),
		Weight:           weight,
		Kind:             in.GetKind(),
		Anonymous:        in.GetAnonymous(),
		ApiKeySecretRefs: apiKeySecretRefs,
	}, nil
}

func ipFraudSecretRefsFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPFraudProviderSettings, providerID string) ([]*commonv1.SecretRef, error) {
	if refs := kernel.CleanIPFraudSecretRefs(in.GetApiKeySecretRefs()); len(refs) > 0 {
		return refs, nil
	}
	rawValues := appcore.CleanList(in.GetApiKeyValues())
	if len(rawValues) == 0 {
		return nil, nil
	}
	out := make([]*commonv1.SecretRef, 0, len(rawValues))
	for index, raw := range rawValues {
		secretID := secretref.StableID("proxy-runtime-ip-fraud-api-key", fmt.Sprintf("%d", in.GetKind()), providerID, fmt.Sprintf("%d", index))
		saved, err := WriteRuntimeSecret(ctx, writer, raw, secretID, kernel.IPFraudAPIKeyPurpose)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}
	return kernel.CleanIPFraudSecretRefs(out), nil
}

func validateIPFraudProvider(provider *proxyruntimev1.ProxyIPFraudProviderSettings, index int, registry *ipfraud.Registry) error {
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return fmt.Errorf("ip_fraud_providers[%d].kind is required", index)
	}
	if provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 {
		return fmt.Errorf("ip_fraud_providers[%d] must use anonymous or api key mode, not both", index)
	}
	if provider.GetAnonymous() && !plugin.SupportsAnonymous() {
		return fmt.Errorf("ip_fraud_providers[%d] does not support anonymous mode", index)
	}
	if !provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 && !plugin.SupportsAPIKey() {
		return fmt.Errorf("ip_fraud_providers[%d] does not support api key mode", index)
	}
	return nil
}
