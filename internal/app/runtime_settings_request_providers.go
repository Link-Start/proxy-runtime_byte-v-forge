package app

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func ipFraudProvidersFromRequest(ctx context.Context, writer secretref.Writer, req []*proxyruntimev1.ProxyIPFraudProviderSettings, current *runtimeSettingsFile, registry *ipfraud.Registry) ([]*proxyruntimev1.ProxyIPFraudProviderSettings, error) {
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

func ipGeoProvidersFromRequest(ctx context.Context, writer secretref.Writer, req []*proxyruntimev1.ProxyIPGeoProviderSettings, current *runtimeSettingsFile, registry *ipgeo.Registry) ([]*proxyruntimev1.ProxyIPGeoProviderSettings, error) {
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

func dynamicIPProvidersFromRequest(req []*proxyruntimev1.ProxyDynamicIPProviderSettings, registry *providerregistry.Registry) ([]*proxyruntimev1.ProxyDynamicIPProviderSettings, error) {
	seenProviders := map[string]struct{}{}
	out := make([]*proxyruntimev1.ProxyDynamicIPProviderSettings, 0, len(req))
	for index, provider := range req {
		item := dynamicIPProviderFromProto(provider)
		if err := validateDynamicIPProvider(item, index, registry); err != nil {
			return nil, err
		}
		id := dynamicIPProviderID(item)
		if _, exists := seenProviders[id]; exists {
			return nil, fmt.Errorf("dynamic_ip_providers[%d] duplicates dynamic provider %q", index, id)
		}
		seenProviders[id] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}
