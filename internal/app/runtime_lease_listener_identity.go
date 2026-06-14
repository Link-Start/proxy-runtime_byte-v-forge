package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func proxyRouteUsername(accountID string) string {
	username := runtimeSafeID(accountID)
	if username == "" {
		username = shortHash(accountID)
	}
	return "acct-" + username
}

func playgroundIngressRule(settings *runtimeSettingsFile) *proxyruntimev1.ProxyIngressRuleSettings {
	for _, rule := range settings.GetIngressRules() {
		if rule.GetRuleId() == playgroundRuleID || strings.TrimSpace(rule.GetUsername()) == playgroundUsername {
			return rule
		}
	}
	return nil
}
