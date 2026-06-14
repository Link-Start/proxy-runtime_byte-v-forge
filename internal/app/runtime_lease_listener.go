package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
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
	active, err := r.store.ListActiveLeaseFacts(ctx, leaseapp.MaxListLimit)
	if err != nil {
		return nil, err
	}
	cleanupPending, err := r.store.CleanupPendingLeaseFacts(ctx)
	if err != nil {
		return nil, err
	}
	return listenerReservedLeaseFacts(active, cleanupPending), nil
}

func listenerReservedLeaseFacts(active []*proxyruntimev1.ProxyDynamicLease, cleanupPending []*proxyruntimev1.ProxyDynamicLease) []*proxyruntimev1.ProxyDynamicLease {
	out := make([]*proxyruntimev1.ProxyDynamicLease, 0, len(active)+len(cleanupPending))
	seen := map[string]struct{}{}
	appendReserved := func(lease *proxyruntimev1.ProxyDynamicLease) {
		if lease.GetListener() == nil {
			return
		}
		leaseID := strings.TrimSpace(lease.GetLeaseId())
		if leaseID != "" {
			if _, exists := seen[leaseID]; exists {
				return
			}
			seen[leaseID] = struct{}{}
		}
		out = append(out, lease)
	}
	for _, lease := range active {
		if leaseapp.HasActiveStatus(lease) {
			appendReserved(lease)
		}
	}
	for _, lease := range cleanupPending {
		if leaseapp.RouteCleanupPending(lease) {
			appendReserved(lease)
		}
	}
	return out
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
