package application

import (
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/proto"
)

func ChangedInUserConnectionUsernames(before *proxyruntimev1.ProxyRuntimePersistentSettings, after *proxyruntimev1.ProxyRuntimePersistentSettings) []string {
	beforeRules := ingressRulesByID(before.GetIngressRules())
	afterRules := ingressRulesByID(after.GetIngressRules())
	beforeProfiles := egressProfilesByID(before.GetEgressProfiles())
	afterProfiles := egressProfilesByID(after.GetEgressProfiles())
	changed := map[string]struct{}{}
	for id, beforeRule := range beforeRules {
		afterRule := afterRules[id]
		if inUserRouteChanged(beforeRule, afterRule, beforeProfiles, afterProfiles) {
			addInUserConnectionUsername(changed, beforeRule.GetUsername())
			if afterRule != nil {
				addInUserConnectionUsername(changed, afterRule.GetUsername())
			}
		}
	}
	for id, afterRule := range afterRules {
		if _, exists := beforeRules[id]; exists {
			continue
		}
		addInUserConnectionUsername(changed, afterRule.GetUsername())
	}
	return sortedKeys(changed)
}

func ingressRulesByID(rules []*proxyruntimev1.ProxyIngressRuleSettings) map[string]*proxyruntimev1.ProxyIngressRuleSettings {
	out := map[string]*proxyruntimev1.ProxyIngressRuleSettings{}
	for _, rule := range rules {
		if id := strings.TrimSpace(rule.GetRuleId()); id != "" {
			out[id] = rule
		}
	}
	return out
}

func egressProfilesByID(profiles []*proxyruntimev1.EgressProfileSettings) map[string]*proxyruntimev1.EgressProfileSettings {
	out := map[string]*proxyruntimev1.EgressProfileSettings{}
	for _, profile := range profiles {
		if id := strings.TrimSpace(profile.GetProfileId()); id != "" {
			out[id] = profile
		}
	}
	return out
}

func inUserRouteChanged(before *proxyruntimev1.ProxyIngressRuleSettings, after *proxyruntimev1.ProxyIngressRuleSettings, beforeProfiles map[string]*proxyruntimev1.EgressProfileSettings, afterProfiles map[string]*proxyruntimev1.EgressProfileSettings) bool {
	if before == nil || after == nil {
		return true
	}
	if before.GetEnabled() != after.GetEnabled() || before.GetUsername() != after.GetUsername() || before.GetPasswordValue() != after.GetPasswordValue() || before.GetProfileId() != after.GetProfileId() {
		return true
	}
	profileID := strings.TrimSpace(after.GetProfileId())
	return !proto.Equal(beforeProfiles[profileID], afterProfiles[profileID])
}

func addInUserConnectionUsername(values map[string]struct{}, username string) {
	username = strings.TrimSpace(username)
	if username != "" {
		values[username] = struct{}{}
	}
}

func sortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
