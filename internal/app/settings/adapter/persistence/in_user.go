package persistence

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

func (s *Store) UpdateInUserRules(ctx context.Context, profiles []*proxygatewayv1.EgressProfileSettings, rules []*proxygatewayv1.ProxyIngressRuleSettings) (*proxygatewayv1.ProxyGatewaySettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
		nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
		if err != nil {
			return nil, err
		}
		dynamicProviderEndpoints := settingsdomain.EnabledDynamicProviderEndpointIDs(settings)
		nextProfiles, err := settingsdomain.EgressProfilesFromRequest(profiles, nativeResourceIDs, dynamicProviderEndpoints)
		if err != nil {
			return nil, err
		}
		nextRules, err := settingsdomain.IngressRulesFromRequest(rules, nextProfiles)
		if err != nil {
			return nil, err
		}
		if err := settingsdomain.RejectOmittedIngressRules(settings.GetIngressRules(), nextRules); err != nil {
			return nil, err
		}
		settingsdomain.ApplyInUserSessionLabels(nextProfiles, nextRules)
		settings.EgressProfiles = nextProfiles
		settings.IngressRules = nextRules
		return settings, nil
	})
}
