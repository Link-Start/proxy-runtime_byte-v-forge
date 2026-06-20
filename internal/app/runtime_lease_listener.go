package app

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) leaseListener(ctx context.Context, settings *runtimeSettingsFile, accountID string, leaseID string) (leaseapp.Listener, error) {
	_ = ctx
	leaseID = appcore.FirstNonEmpty(leaseID, accountID)
	listener, err := leaseapp.NewDynamicListener(leaseapp.DynamicListenerInput{
		ID:                  "lease-" + appcore.ShortHash(leaseID),
		Addr:                r.cfg.LocalAddr,
		Protocol:            r.cfg.LocalProtocol,
		Route:               config.ListenerRouteProvider,
		AccountID:           accountID,
		LeaseID:             leaseID,
		DefaultUsername:     proxyRouteUsername(leaseID),
		FallbackPassword:    r.cfg.LocalPassword,
		IngressRules:        settings.GetIngressRules(),
		PlaygroundAccountID: kernel.PlaygroundProfileID,
		PlaygroundRuleID:    kernel.PlaygroundRuleID,
		PlaygroundUsername:  kernel.PlaygroundUsername,
	})
	if errors.Is(err, leaseapp.ErrListenerPasswordRequired) {
		return leaseapp.Listener{}, appcore.FailedPrecondition(err.Error(), nil)
	}
	return listener, err
}

func proxyRouteUsername(accountID string) string {
	username := appcore.RuntimeSafeID(accountID)
	if username == "" {
		username = appcore.ShortHash(accountID)
	}
	return "acct-" + username
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
	return leaseapp.ReservedListenerLeaseFacts(active, cleanupPending), nil
}
