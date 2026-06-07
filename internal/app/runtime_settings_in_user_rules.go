package app

import (
	"context"
	"strings"

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
	applyInUserSessionLabels(nextProfiles, nextRules)
	settings.EgressProfiles = nextProfiles
	settings.IngressRules = nextRules
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func applyInUserSessionLabels(profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) {
	profilesByID := map[string]*proxyruntimev1.EgressProfileSettings{}
	for _, profile := range profiles {
		profilesByID[runtimeSafeID(profile.GetProfileId())] = profile
	}
	for _, rule := range rules {
		sessionID := inUserSessionID(rule.GetUsername())
		if sessionID == "" {
			continue
		}
		profile := profilesByID[runtimeSafeID(rule.GetProfileId())]
		if profile == nil || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		policy := profile.GetExit().GetDynamicIpPolicy()
		if policy.GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
			continue
		}
		if policy.Labels == nil {
			policy.Labels = map[string]string{}
		}
		policy.Labels["session_id"] = sessionID
	}
}

func inUserSessionID(username string) string {
	_, session, ok := strings.Cut(strings.TrimSpace(username), "-session-")
	if !ok {
		return ""
	}
	return runtimeSafeID(session)
}
