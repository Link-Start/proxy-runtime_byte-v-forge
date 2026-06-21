package persistence

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

func (s *Store) UpdateEgressProfiles(ctx context.Context, profiles []*proxygatewayv1.EgressProfileSettings) (*proxygatewayv1.ProxyGatewaySettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
		nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
		if err != nil {
			return nil, err
		}
		nextProfiles, err := settingsdomain.EgressProfilesFromRequest(profiles, nativeResourceIDs, settingsdomain.EnabledDynamicProviderEndpointIDs(settings))
		if err != nil {
			return nil, err
		}
		nextRules, err := settingsdomain.IngressRulesFromRequest(settings.GetIngressRules(), nextProfiles)
		if err != nil {
			return nil, err
		}
		settingsdomain.ApplyInUserSessionLabels(nextProfiles, nextRules)
		settings.EgressProfiles = nextProfiles
		settings.IngressRules = nextRules
		return settings, nil
	})
}

func (s *Store) UpdateIngressRules(ctx context.Context, rules []*proxygatewayv1.ProxyIngressRuleSettings) (*proxygatewayv1.ProxyGatewaySettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
		nextRules, err := settingsdomain.IngressRulesFromRequest(rules, settings.GetEgressProfiles())
		if err != nil {
			return nil, err
		}
		settingsdomain.ApplyInUserSessionLabels(settings.EgressProfiles, nextRules)
		settings.IngressRules = nextRules
		return settings, nil
	})
}
