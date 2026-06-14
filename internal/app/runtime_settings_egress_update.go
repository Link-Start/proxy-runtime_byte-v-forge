package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *runtimeSettingsStore) updateEgressProfiles(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
		nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
		if err != nil {
			return nil, err
		}
		nextProfiles, err := egressProfilesFromRequest(profiles, nativeResourceIDs, enabledDynamicProviderEndpointIDs(settings))
		if err != nil {
			return nil, err
		}
		nextRules, err := ingressRulesFromRequest(settings.GetIngressRules(), nextProfiles)
		if err != nil {
			return nil, err
		}
		applyInUserSessionLabels(nextProfiles, nextRules)
		settings.EgressProfiles = nextProfiles
		settings.IngressRules = nextRules
		return settings, nil
	})
}

func (s *runtimeSettingsStore) updateIngressRules(ctx context.Context, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
		nextRules, err := ingressRulesFromRequest(rules, settings.GetEgressProfiles())
		if err != nil {
			return nil, err
		}
		applyInUserSessionLabels(settings.EgressProfiles, nextRules)
		settings.IngressRules = nextRules
		return settings, nil
	})
}
