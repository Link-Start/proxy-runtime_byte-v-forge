package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func ListenerPassword(rules []*proxyruntimev1.ProxyIngressRuleSettings, profileID string, fallback string) string {
	if password := strings.TrimSpace(fallback); password != "" {
		return password
	}
	if rule := IngressRuleForProfile(rules, profileID); rule != nil {
		return rule.GetPasswordValue()
	}
	return ""
}

func IngressRuleByUsername(rules []*proxyruntimev1.ProxyIngressRuleSettings, username string) *proxyruntimev1.ProxyIngressRuleSettings {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}
	for _, rule := range rules {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetUsername()) != username || strings.TrimSpace(rule.GetProfileId()) == "" {
			continue
		}
		return rule
	}
	return nil
}

func IngressRuleForProfile(rules []*proxyruntimev1.ProxyIngressRuleSettings, profileID string) *proxyruntimev1.ProxyIngressRuleSettings {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil
	}
	for _, rule := range rules {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetProfileId()) != profileID || strings.TrimSpace(rule.GetPasswordValue()) == "" {
			continue
		}
		return rule
	}
	return nil
}

func PlaygroundIngressRule(rules []*proxyruntimev1.ProxyIngressRuleSettings, ruleID string, username string) *proxyruntimev1.ProxyIngressRuleSettings {
	ruleID = strings.TrimSpace(ruleID)
	username = strings.TrimSpace(username)
	for _, rule := range rules {
		if ruleID != "" && rule.GetRuleId() == ruleID {
			return rule
		}
		if username != "" && strings.TrimSpace(rule.GetUsername()) == username {
			return rule
		}
	}
	return nil
}
