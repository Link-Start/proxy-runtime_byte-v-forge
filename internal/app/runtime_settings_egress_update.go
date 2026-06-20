package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	settingsdomain "github.com/byte-v-forge/proxy-runtime/internal/app/settings/domain"
)

func (s *runtimeSettingsStore) updateEgressProfiles(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
		nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
		if err != nil {
			return nil, err
		}
		nextProfiles, err := settingsdomain.EgressProfilesFromRequest(profiles, nativeResourceIDs, enabledDynamicProviderEndpointIDs(settings))
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

func (s *runtimeSettingsStore) updateIngressRules(ctx context.Context, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
		nextRules, err := settingsdomain.IngressRulesFromRequest(rules, settings.GetEgressProfiles())
		if err != nil {
			return nil, err
		}
		settingsdomain.ApplyInUserSessionLabels(settings.EgressProfiles, nextRules)
		settings.IngressRules = nextRules
		return settings, nil
	})
}
