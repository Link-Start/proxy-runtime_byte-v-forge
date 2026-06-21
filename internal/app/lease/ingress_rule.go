package lease

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func ListenerPassword(rules []*proxygatewayv1.ProxyIngressRuleSettings, profileID string, fallback string) string {
	if password := strings.TrimSpace(fallback); password != "" {
		return password
	}
	if rule := IngressRuleForProfile(rules, profileID); rule != nil {
		return rule.GetPasswordValue()
	}
	return ""
}

func IngressRuleByUsername(rules []*proxygatewayv1.ProxyIngressRuleSettings, username string) *proxygatewayv1.ProxyIngressRuleSettings {
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

func IngressRuleForProfile(rules []*proxygatewayv1.ProxyIngressRuleSettings, profileID string) *proxygatewayv1.ProxyIngressRuleSettings {
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

func PlaygroundIngressRule(rules []*proxygatewayv1.ProxyIngressRuleSettings, ruleID string, username string) *proxygatewayv1.ProxyIngressRuleSettings {
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
