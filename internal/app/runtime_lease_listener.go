package app

import (
	"context"
	"strings"

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
