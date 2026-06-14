package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func leaseListenerPassword(settings *runtimeSettingsFile, profileID string, fallback string) string {
	if password := strings.TrimSpace(fallback); password != "" {
		return password
	}
	if rule := ingressRuleForProfile(settings, profileID); rule != nil {
		return rule.GetPasswordValue()
	}
	return ""
}

func ingressRuleForProfile(settings *runtimeSettingsFile, profileID string) *proxyruntimev1.ProxyIngressRuleSettings {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" || settings == nil {
		return nil
	}
	for _, rule := range settings.GetIngressRules() {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetProfileId()) != profileID || strings.TrimSpace(rule.GetPasswordValue()) == "" {
			continue
		}
		return rule
	}
	return nil
}
