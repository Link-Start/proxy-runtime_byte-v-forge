package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *runtimeSettingsStore) updateInUserRules(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
	if err != nil {
		return nil, err
	}
	dynamicProviderEndpoints := enabledDynamicProviderEndpointIDs(settings)
	nextProfiles, err := egressProfilesFromRequest(profiles, nativeResourceIDs, dynamicProviderEndpoints)
	if err != nil {
		return nil, err
	}
	nextRules, err := ingressRulesFromRequest(rules, nextProfiles)
	if err != nil {
		return nil, err
	}
	settings.EgressProfiles = nextProfiles
	settings.IngressRules = nextRules
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}
