package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) leaseListener(ctx context.Context, settings *runtimeSettingsFile, accountID string, leaseID string) (config.EgressListener, error) {
	_ = ctx
	leaseID = firstNonEmpty(leaseID, accountID)
	id := "lease-" + shortHash(leaseID)
	username := proxyRouteUsername(leaseID)
	password := leaseListenerPassword(settings, accountID, r.cfg.LocalPassword)
	if accountID == playgroundProfileID {
		if rule := playgroundIngressRule(settings); rule != nil {
			username = rule.GetUsername()
			password = firstNonEmpty(rule.GetPasswordValue(), password)
		}
	}
	if strings.TrimSpace(password) == "" {
		return config.EgressListener{}, failedPrecondition("dynamic lease listener password is not configured", nil)
	}
	return config.EgressListener{
		ID:       id,
		Addr:     r.cfg.LocalAddr,
		Protocol: r.cfg.LocalProtocol,
		Route:    config.ListenerRouteProvider,
		Username: username,
		Password: password,
		Labels: map[string]string{
			"mode":       "dynamic_ip_session_lease",
			"account_id": accountID,
			"lease_id":   leaseID,
		},
	}, nil
}

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

func (r *Runtime) listenerReservedLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	leases, err := r.store.ListLeaseFacts(ctx, true)
	if err != nil {
		return nil, err
	}
	out := make([]*proxyruntimev1.ProxyDynamicLease, 0, len(leases))
	for _, lease := range leases {
		if lease.GetListener() == nil {
			continue
		}
		if lease.GetStatus() == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE || leaseRouteCleanupPending(lease) {
			out = append(out, lease)
		}
	}
	return out, nil
}

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
