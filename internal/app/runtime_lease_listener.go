package app

import (
	"context"
	"strings"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) leaseListener(ctx context.Context, settings *runtimeSettingsFile, accountID string, leaseID string) (leaseapp.Listener, error) {
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
		return leaseapp.Listener{}, failedPrecondition("dynamic lease listener password is not configured", nil)
	}
	return leaseapp.Listener{
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
