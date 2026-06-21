package secret

import (
	"context"
	"errors"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"
	"github.com/byte-v-forge/proxy-gateway/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-gateway/internal/provider/registry"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

// SettingsFromRequest builds the persisted runtime settings from an update
// request, writing inline secrets through writer and validating against the
// provider registries.
func SettingsFromRequest(ctx context.Context, writer secretref.Writer, req *proxygatewayv1.UpdateProxyGatewaySettingsRequest, current *proxygatewayv1.ProxyGatewayPersistentSettings, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry, nativeResourceIDs map[string]struct{}) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
	current = kernel.NormalizeRuntimeSettingsWithProviders(current, ipFraudProviders, ipGeoProviders)
	edgeCanary, err := edgeCanaryFromRequest(ctx, writer, req.GetEdgeCanary(), current.GetEdgeCanary())
	if err != nil {
		return nil, err
	}
	settings := &proxygatewayv1.ProxyGatewayPersistentSettings{
		EdgeCanary:         edgeCanary,
		IpFraudProviders:   make([]*proxygatewayv1.ProxyIPFraudProviderSettings, 0, len(req.GetIpFraudProviders())),
		IpGeoProviders:     make([]*proxygatewayv1.ProxyIPGeoProviderSettings, 0, len(req.GetIpGeoProviders())),
		DynamicIpProviders: make([]*proxygatewayv1.ProxyDynamicIPProviderSettings, 0, len(req.GetDynamicIpProviders())),
		EgressProfiles:     make([]*proxygatewayv1.EgressProfileSettings, 0, len(req.GetEgressProfiles())),
		IngressRules:       make([]*proxygatewayv1.ProxyIngressRuleSettings, 0, len(req.GetIngressRules())),
		CheckSettings:      kernel.CheckSettingsFromRequest(req.GetCheckSettings(), current.GetCheckSettings()),
	}
	if settingsdomain.EdgeCanaryEnabled(settings.GetEdgeCanary()) && strings.TrimSpace(settings.GetEdgeCanary().GetUrl()) == "" {
		return nil, errors.New("edge canary url is required when enabled")
	}
	settings.IpFraudProviders, err = IPFraudProvidersFromRequest(ctx, writer, req.GetIpFraudProviders(), current, ipFraudProviders)
	if err != nil {
		return nil, err
	}
	settings.IpGeoProviders, err = IPGeoProvidersFromRequest(ctx, writer, req.GetIpGeoProviders(), current, ipGeoProviders)
	if err != nil {
		return nil, err
	}
	settings.DynamicIpProviders, err = settingsdomain.DynamicIPProvidersFromRequest(req.GetDynamicIpProviders(), accountProviders)
	if err != nil {
		return nil, err
	}
	dynamicProviderEndpoints := settingsdomain.EnabledDynamicProviderEndpointIDs(settings)
	settings.EgressProfiles, err = settingsdomain.EgressProfilesFromRequest(req.GetEgressProfiles(), nativeResourceIDs, dynamicProviderEndpoints)
	if err != nil {
		return nil, err
	}
	settings.IngressRules, err = settingsdomain.IngressRulesFromRequest(req.GetIngressRules(), settings.GetEgressProfiles())
	if err != nil {
		return nil, err
	}
	return kernel.NormalizeRuntimeSettingsWithProviders(settings, ipFraudProviders, ipGeoProviders), nil
}
