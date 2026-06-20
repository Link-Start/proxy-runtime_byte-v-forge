package domain

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func cloneIngressRule(in *proxyruntimev1.ProxyIngressRuleSettings) *proxyruntimev1.ProxyIngressRuleSettings {
	return kernel.IngressRuleFromProto(in, 0)
}

func IngressRulesFromRequest(in []*proxyruntimev1.ProxyIngressRuleSettings, profiles []*proxyruntimev1.EgressProfileSettings) ([]*proxyruntimev1.ProxyIngressRuleSettings, error) {
	out := make([]*proxyruntimev1.ProxyIngressRuleSettings, 0, len(in))
	seenIDs := map[string]struct{}{}
	seenUsers := map[string]struct{}{}
	enabledProfiles := enabledEgressProfileIDsFromProfiles(profiles)
	for index, rule := range in {
		item := kernel.IngressRuleFromProto(rule, index)
		if err := validateIngressRule(item, index, enabledProfiles); err != nil {
			return nil, err
		}
		if _, exists := seenIDs[item.GetRuleId()]; exists {
			return nil, fmt.Errorf("ingress_rules[%d] duplicates rule %q", index, item.GetRuleId())
		}
		seenIDs[item.GetRuleId()] = struct{}{}
		if username := item.GetUsername(); username != "" {
			if _, exists := seenUsers[username]; exists {
				return nil, fmt.Errorf("ingress_rules[%d] duplicates username %q", index, username)
			}
			seenUsers[username] = struct{}{}
		}
		out = append(out, item)
	}
	return out, nil
}

func validateIngressRule(rule *proxyruntimev1.ProxyIngressRuleSettings, index int, enabledProfiles map[string]struct{}) error {
	if rule.GetRuleId() == "" {
		return fmt.Errorf("ingress_rules[%d].rule_id is required", index)
	}
	if !rule.GetEnabled() {
		return nil
	}
	if rule.GetUsername() == "" {
		return fmt.Errorf("ingress_rules[%d].username is required when enabled", index)
	}
	profileID := rule.GetProfileId()
	if profileID == "" {
		return fmt.Errorf("ingress_rules[%d].profile_id is required when enabled", index)
	}
	if _, exists := enabledProfiles[profileID]; !exists {
		return fmt.Errorf("ingress_rules[%d].profile_id %q is not enabled", index, profileID)
	}
	return nil
}
