package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/secretref"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

func settingsFromRequest(ctx context.Context, writer secretref.Writer, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest, current *runtimeSettingsFile, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry, nativeResourceIDs map[string]struct{}) (*runtimeSettingsFile, error) {
	current = normalizeRuntimeSettingsWithProviders(current, ipFraudProviders, ipGeoProviders)
	edgeCanary, err := edgeCanaryFromRequest(ctx, writer, req.GetEdgeCanary(), current.GetEdgeCanary())
	if err != nil {
		return nil, err
	}
	settings := &proxyruntimev1.ProxyRuntimePersistentSettings{
		EdgeCanary:         edgeCanary,
		IpFraudProviders:   make([]*proxyruntimev1.ProxyIPFraudProviderSettings, 0, len(req.GetIpFraudProviders())),
		IpGeoProviders:     make([]*proxyruntimev1.ProxyIPGeoProviderSettings, 0, len(req.GetIpGeoProviders())),
		DynamicIpProviders: make([]*proxyruntimev1.ProxyDynamicIPProviderSettings, 0, len(req.GetDynamicIpProviders())),
		EgressProfiles:     make([]*proxyruntimev1.EgressProfileSettings, 0, len(req.GetEgressProfiles())),
		IngressRules:       make([]*proxyruntimev1.ProxyIngressRuleSettings, 0, len(req.GetIngressRules())),
		CheckSettings:      checkSettingsFromRequest(req.GetCheckSettings(), current.GetCheckSettings()),
	}
	if edgeCanaryEnabled(settings.GetEdgeCanary()) && strings.TrimSpace(settings.GetEdgeCanary().GetUrl()) == "" {
		return nil, errors.New("edge canary url is required when enabled")
	}
	currentProviders := ipFraudProviderSecrets(current, ipFraudProviders)
	seenProviders := map[string]struct{}{}
	for index, provider := range req.GetIpFraudProviders() {
		item, err := ipFraudProviderFromRequest(ctx, writer, provider, currentProviders, index, ipFraudProviders)
		if err != nil {
			return nil, err
		}
		if err := validateIPFraudProvider(item, index, ipFraudProviders); err != nil {
			return nil, err
		}
		key := providerSecretKey(item.GetKind(), item.GetProviderId())
		if _, exists := seenProviders[key]; exists {
			return nil, fmt.Errorf("ip_fraud_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seenProviders[key] = struct{}{}
		settings.IpFraudProviders = append(settings.IpFraudProviders, item)
	}
	currentGeoProviders := ipGeoProviderSecrets(current, ipGeoProviders)
	seenGeoProviders := map[string]struct{}{}
	for index, provider := range req.GetIpGeoProviders() {
		item, err := ipGeoProviderFromRequest(ctx, writer, provider, currentGeoProviders, index, ipGeoProviders)
		if err != nil {
			return nil, err
		}
		if err := validateIPGeoProvider(item, index, ipGeoProviders); err != nil {
			return nil, err
		}
		key := ipGeoProviderSecretKey(item.GetKind(), item.GetProviderId())
		if _, exists := seenGeoProviders[key]; exists {
			return nil, fmt.Errorf("ip_geo_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seenGeoProviders[key] = struct{}{}
		settings.IpGeoProviders = append(settings.IpGeoProviders, item)
	}
	seenDynamicProviders := map[string]struct{}{}
	for index, provider := range req.GetDynamicIpProviders() {
		item := dynamicIPProviderFromProto(provider)
		if err := validateDynamicIPProvider(item, index, accountProviders); err != nil {
			return nil, err
		}
		id := dynamicIPProviderID(item)
		if _, exists := seenDynamicProviders[id]; exists {
			return nil, fmt.Errorf("dynamic_ip_providers[%d] duplicates dynamic provider %q", index, id)
		}
		seenDynamicProviders[id] = struct{}{}
		settings.DynamicIpProviders = append(settings.DynamicIpProviders, item)
	}
	dynamicProviderEndpoints := enabledDynamicProviderEndpointIDs(settings)
	settings.EgressProfiles, err = egressProfilesFromRequest(req.GetEgressProfiles(), nativeResourceIDs, dynamicProviderEndpoints)
	if err != nil {
		return nil, err
	}
	settings.IngressRules, err = ingressRulesFromRequest(req.GetIngressRules(), settings.GetEgressProfiles())
	if err != nil {
		return nil, err
	}
	return normalizeRuntimeSettingsWithProviders(settings, ipFraudProviders, ipGeoProviders), nil
}
