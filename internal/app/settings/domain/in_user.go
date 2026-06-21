package domain

import (
	"sort"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func ApplyInUserSessionLabels(profiles []*proxygatewayv1.EgressProfileSettings, rules []*proxygatewayv1.ProxyIngressRuleSettings) {
	profilesByID := map[string]*proxygatewayv1.EgressProfileSettings{}
	for _, profile := range profiles {
		profilesByID[appcore.RuntimeSafeID(profile.GetProfileId())] = profile
	}
	for _, rule := range rules {
		sessionID := inUserSessionID(rule.GetUsername())
		if sessionID == "" {
			continue
		}
		profile := profilesByID[appcore.RuntimeSafeID(rule.GetProfileId())]
		if profile == nil || profile.GetExit().GetKind() != proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		policy := profile.GetExit().GetDynamicIpPolicy()
		if policy.GetMode() != proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
			continue
		}
		if policy.Labels == nil {
			policy.Labels = map[string]string{}
		}
		policy.Labels["session_id"] = sessionID
	}
}

func RejectOmittedIngressRules(current []*proxygatewayv1.ProxyIngressRuleSettings, next []*proxygatewayv1.ProxyIngressRuleSettings) error {
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

func existingIngressRuleIDs(rules []*proxygatewayv1.ProxyIngressRuleSettings) map[string]struct{} {
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
