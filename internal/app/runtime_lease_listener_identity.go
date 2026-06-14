package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func proxyRouteUsername(accountID string) string {
	username := runtimeSafeID(accountID)
	if username == "" {
		username = shortHash(accountID)
	}
	return "acct-" + username
}

func playgroundIngressRule(settings *runtimeSettingsFile) *proxyruntimev1.ProxyIngressRuleSettings {
	return leaseapp.PlaygroundIngressRule(settings.GetIngressRules(), playgroundRuleID, playgroundUsername)
}
