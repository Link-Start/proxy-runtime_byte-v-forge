package app

import (
	"context"
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func settingsFromRequest(ctx context.Context, writer secretref.Writer, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest, current *runtimeSettingsFile, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry, nativeResourceIDs map[string]struct{}) (*runtimeSettingsFile, error) {
	current = settingscore.NormalizeRuntimeSettingsWithProviders(current, ipFraudProviders, ipGeoProviders)
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
		CheckSettings:      settingscore.CheckSettingsFromRequest(req.GetCheckSettings(), current.GetCheckSettings()),
	}
	if edgeCanaryEnabled(settings.GetEdgeCanary()) && strings.TrimSpace(settings.GetEdgeCanary().GetUrl()) == "" {
		return nil, errors.New("edge canary url is required when enabled")
	}
	settings.IpFraudProviders, err = settingscore.IPFraudProvidersFromRequest(ctx, writer, req.GetIpFraudProviders(), current, ipFraudProviders)
	if err != nil {
		return nil, err
	}
	settings.IpGeoProviders, err = settingscore.IPGeoProvidersFromRequest(ctx, writer, req.GetIpGeoProviders(), current, ipGeoProviders)
	if err != nil {
		return nil, err
	}
	settings.DynamicIpProviders, err = dynamicIPProvidersFromRequest(req.GetDynamicIpProviders(), accountProviders)
	if err != nil {
		return nil, err
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
	return settingscore.NormalizeRuntimeSettingsWithProviders(settings, ipFraudProviders, ipGeoProviders), nil
}
