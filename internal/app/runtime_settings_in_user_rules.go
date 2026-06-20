package app

import (
	"context"
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (s *runtimeSettingsStore) updateInUserRules(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
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
		if err := rejectOmittedIngressRules(settings.GetIngressRules(), nextRules); err != nil {
			return nil, err
		}
		applyInUserSessionLabels(nextProfiles, nextRules)
		settings.EgressProfiles = nextProfiles
		settings.IngressRules = nextRules
		return settings, nil
	})
}

func applyInUserSessionLabels(profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) {
	profilesByID := map[string]*proxyruntimev1.EgressProfileSettings{}
	for _, profile := range profiles {
		profilesByID[appcore.RuntimeSafeID(profile.GetProfileId())] = profile
	}
	for _, rule := range rules {
		sessionID := inUserSessionID(rule.GetUsername())
		if sessionID == "" {
			continue
		}
		profile := profilesByID[appcore.RuntimeSafeID(rule.GetProfileId())]
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

func rejectOmittedIngressRules(current []*proxyruntimev1.ProxyIngressRuleSettings, next []*proxyruntimev1.ProxyIngressRuleSettings) error {
	missing := existingIngressRuleIDs(current)
	for _, rule := range next {
		delete(missing, appcore.RuntimeSafeID(rule.GetRuleId()))
	}
	if len(missing) == 0 {
		return nil
	}
	ids := make([]string, 0, len(missing))
	for id := range missing {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return appcore.FailedPrecondition("in-user update omitted existing ingress rules: "+strings.Join(ids, ", "), nil)
}

func existingIngressRuleIDs(rules []*proxyruntimev1.ProxyIngressRuleSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, rule := range rules {
		if id := appcore.RuntimeSafeID(rule.GetRuleId()); id != "" {
			out[id] = struct{}{}
		}
	}
	return out
}

func inUserSessionID(username string) string {
	_, session, ok := strings.Cut(strings.TrimSpace(username), "-session-")
	if !ok {
		return ""
	}
	return appcore.RuntimeSafeID(session)
}
