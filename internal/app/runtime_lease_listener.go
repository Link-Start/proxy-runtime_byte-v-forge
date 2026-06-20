package app

import (
	"context"
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (r *Runtime) leaseListener(ctx context.Context, settings *runtimeSettingsFile, accountID string, leaseID string) (leaseapp.Listener, error) {
	_ = ctx
	leaseID = firstNonEmpty(leaseID, accountID)
	listener, err := leaseapp.NewDynamicListener(leaseapp.DynamicListenerInput{
		ID:                  "lease-" + shortHash(leaseID),
		Addr:                r.cfg.LocalAddr,
		Protocol:            r.cfg.LocalProtocol,
		Route:               config.ListenerRouteProvider,
		AccountID:           accountID,
		LeaseID:             leaseID,
		DefaultUsername:     proxyRouteUsername(leaseID),
		FallbackPassword:    r.cfg.LocalPassword,
		IngressRules:        settings.GetIngressRules(),
		PlaygroundAccountID: playgroundProfileID,
		PlaygroundRuleID:    playgroundRuleID,
		PlaygroundUsername:  playgroundUsername,
	})
	if errors.Is(err, leaseapp.ErrListenerPasswordRequired) {
		return leaseapp.Listener{}, appcore.FailedPrecondition(err.Error(), nil)
	}
	return listener, err
}
