package app

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/secretref"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
)

func providerSecrets(settings *runtimeSettingsFile, providers *ipfraud.Registry) map[string][]*commonv1.SecretRef {
	secrets := map[string][]*commonv1.SecretRef{}
	for _, item := range normalizeRuntimeSettingsWithProviders(settings, providers).GetIpFraudProviders() {
		secrets[providerSecretKey(item.GetKind(), item.GetProviderId())] = cleanIPFraudSecretRefs(item.GetApiKeySecretRefs())
	}
	return secrets
}

func providerSecretKey(kind proxyruntimev1.ProxyIPFraudProviderKind, id string) string {
	return fmt.Sprintf("%d:%s", kind, strings.TrimSpace(id))
}

func ipFraudProviders(ctx context.Context, resolver secretref.Resolver, settings *runtimeSettingsFile, registry *ipfraud.Registry) ([]ipfraud.ProviderConfig, error) {
	items := normalizeRuntimeSettingsWithProviders(settings, registry).GetIpFraudProviders()
	providers := make([]ipfraud.ProviderConfig, 0, len(items))
	for _, item := range items {
		auth, err := ipFraudAuth(ctx, resolver, item, registry)
		if err != nil {
			return nil, err
		}
		providers = append(providers, ipfraud.ProviderConfig{
			ID:     item.GetProviderId(),
			Kind:   item.GetKind(),
			Weight: int(item.GetWeight()),
			Auth:   auth,
		})
	}
	return providers, nil
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
		apiKeySecretRefs = current[providerSecretKey(in.GetKind(), id)]
	}
	weight := in.GetWeight()
	if weight == 0 {
		weight = providerDefaultWeight(in.GetKind(), index, registry)
	}
	return &proxyruntimev1.ProxyIPFraudProviderSettings{ProviderId: id, Weight: weight, Kind: in.GetKind(), Anonymous: in.GetAnonymous(), ApiKeySecretRefs: apiKeySecretRefs}, nil
}

func normalizeIPFraudProvider(provider *proxyruntimev1.ProxyIPFraudProviderSettings, index int, registry *ipfraud.Registry) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.ApiKeySecretRefs = cleanIPFraudSecretRefs(provider.GetApiKeySecretRefs())
	if provider.ProviderId == "" {
		provider.ProviderId = registry.DefaultProviderID(provider.GetKind())
	}
	if provider.Weight == 0 {
		provider.Weight = providerDefaultWeight(provider.GetKind(), index, registry)
	}
}

func validateIPFraudProvider(provider *proxyruntimev1.ProxyIPFraudProviderSettings, index int, registry *ipfraud.Registry) error {
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return fmt.Errorf("ip_fraud_providers[%d].kind is required", index)
	}
	if provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) > 0 {
		return fmt.Errorf("ip_fraud_providers[%d] must use anonymous or api_key_secret_refs, not both", index)
	}
	if provider.GetAnonymous() && !plugin.SupportsAnonymous() {
		return fmt.Errorf("ip_fraud_providers[%d] does not support anonymous mode", index)
	}
	if !provider.GetAnonymous() && !plugin.SupportsAPIKey() {
		return fmt.Errorf("ip_fraud_providers[%d] does not support api key mode", index)
	}
	if !provider.GetAnonymous() && len(provider.GetApiKeySecretRefs()) == 0 {
		return fmt.Errorf("ip_fraud_providers[%d].api_key_secret_refs is required when anonymous is false", index)
	}
	return nil
}

func supportedIPFraudProviders(providers []*proxyruntimev1.ProxyIPFraudProviderSettings, registry *ipfraud.Registry) []*proxyruntimev1.ProxyIPFraudProviderSettings {
	out := make([]*proxyruntimev1.ProxyIPFraudProviderSettings, 0, len(providers))
	for _, provider := range providers {
		if registry.IsProviderKindSupported(provider.GetKind()) {
			out = append(out, provider)
		}
	}
	return out
}

func ipFraudAuth(ctx context.Context, resolver secretref.Resolver, provider *proxyruntimev1.ProxyIPFraudProviderSettings, registry *ipfraud.Registry) (ipfraud.AuthConfig, error) {
	if provider.GetAnonymous() {
		return ipfraud.AuthConfig{Anonymous: &ipfraud.AnonymousAuthConfig{}}, nil
	}
	plugin, ok := registry.PluginForKind(provider.GetKind())
	if !ok {
		return ipfraud.AuthConfig{}, nil
	}
	values, err := resolveSecretRefs(ctx, resolver, provider.GetApiKeySecretRefs())
	if err != nil {
		return ipfraud.AuthConfig{}, err
	}
	return plugin.Auth(values, false), nil
}

func cleanIPFraudSecretRefs(values []*commonv1.SecretRef) []*commonv1.SecretRef {
	return cleanSecretRefs(values, "proxy-runtime", "ip_fraud_api_key")
}

func ipFraudSecretRefsFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPFraudProviderSettings, providerID string) ([]*commonv1.SecretRef, error) {
	rawRefs := in.GetApiKeySecretRefs()
	if len(rawRefs) == 0 {
		return nil, nil
	}
	out := make([]*commonv1.SecretRef, 0, len(rawRefs))
	for index, ref := range rawRefs {
		raw := strings.TrimSpace(ref.GetSecretId())
		if raw == "" {
			continue
		}
		secretID := secretref.StableID("proxy-runtime-ip-fraud-api-key", fmt.Sprintf("%d", in.GetKind()), providerID, fmt.Sprintf("%d", index))
		saved, err := writeRuntimeSecret(ctx, writer, raw, secretID, "ip_fraud_api_key")
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}
	return cleanIPFraudSecretRefs(out), nil
}

func resolveSecretRefs(ctx context.Context, resolver secretref.Resolver, refs []*commonv1.SecretRef) ([]string, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	if resolver == nil {
		return nil, fmt.Errorf("secret resolver is required")
	}
	out := make([]string, 0, len(refs))
	for _, ref := range cleanIPFraudSecretRefs(refs) {
		value, err := resolver.ResolveSecret(ctx, ref)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out, nil
}

func defaultProviderWeight(index int) uint32 {
	if index < 0 {
		return 100
	}
	if index > 9 {
		return 10
	}
	return uint32(100 - index*10)
}

func providerDefaultWeight(kind proxyruntimev1.ProxyIPFraudProviderKind, index int, registry *ipfraud.Registry) uint32 {
	if plugin, ok := registry.PluginForKind(kind); ok {
		return plugin.DefaultWeight()
	}
	return defaultProviderWeight(index)
}
