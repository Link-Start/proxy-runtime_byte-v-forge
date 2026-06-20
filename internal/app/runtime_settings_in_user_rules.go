package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	settingsdomain "github.com/byte-v-forge/proxy-runtime/internal/app/settings/domain"
)

func (s *runtimeSettingsStore) updateInUserRules(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
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
