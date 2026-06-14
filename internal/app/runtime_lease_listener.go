package app

import (
	"context"
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) leaseListener(ctx context.Context, settings *runtimeSettingsFile, accountID string, leaseID string) (leaseapp.Listener, error) {
	_ = ctx
	leaseID = firstNonEmpty(leaseID, accountID)
	listener, err := leaseapp.NewListener(leaseapp.ListenerInput{
		ID:        "lease-" + shortHash(leaseID),
		Addr:      r.cfg.LocalAddr,
		Protocol:  r.cfg.LocalProtocol,
		Route:     config.ListenerRouteProvider,
		Username:  leaseListenerUsername(settings, accountID, leaseID),
		Password:  leaseListenerPasswordValue(settings, accountID, r.cfg.LocalPassword),
		AccountID: accountID,
		LeaseID:   leaseID,
	})
	if errors.Is(err, leaseapp.ErrListenerPasswordRequired) {
		return leaseapp.Listener{}, failedPrecondition(err.Error(), nil)
	}
	return listener, err
}
