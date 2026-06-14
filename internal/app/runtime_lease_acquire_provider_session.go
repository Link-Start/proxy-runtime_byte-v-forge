package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireLeaseWithProviderAccountLock(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string) (*proxyruntimev1.ProxyDynamicLease, error) {
	providerCfg, providerAccountID, err := leaseapp.ProviderConfigForGateway(ctx, c.deps.store, providerAccountID, selection.endpoint)
	if err != nil {
		return nil, err
	}
	providerClient, err := c.newSessionProvider(providerCfg)
	if err != nil {
		return nil, invalidArgument("provider account configuration is invalid", err)
	}
	session, nodes, err := leaseapp.CreateAndFetchProviderSession(ctx, providerClient, req, selection.plan, concurrencyHolder)
	switch leaseapp.ClassifyProviderSessionError(session, err) {
	case leaseapp.ProviderSessionCreateError:
		return nil, unavailable("provider session create failed", err)
	case leaseapp.ProviderSessionFetchError:
		failure := c.newFailedAcquireRecorder(req, providerAccountID, providerClient, session, selection.plan)
		failure.BeforeRoute(ctx, "provider session fetch failed")
		return nil, unavailable("provider session fetch failed", err)
	}
	failure := c.newFailedAcquireRecorder(req, providerAccountID, providerClient, session, selection.plan)
	lineBinding, err := leaseapp.PrepareRouteLineBinding(ctx, nodes, req.GetAccountId(), func(ctx context.Context, accountID string) (string, map[string]string, error) {
		return c.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
	})
	if err != nil {
		failure.BeforeRoute(ctx, "lease line resolution failed")
		return nil, err
	}
	var lease *proxyruntimev1.ProxyDynamicLease
	err = leaseapp.WithSessionListenerAllocationLock(ctx, c.deps.locks, func(ctx context.Context) error {
		var err error
		lease, err = c.applyAcquiredLeaseRoute(ctx, acquiredLeaseFlow{
			advertisedHost:    advertisedHost,
			request:           req,
			settings:          settings,
			selection:         selection,
			providerAccountID: providerAccountID,
			leaseID:           leaseID,
			concurrencyHolder: concurrencyHolder,
			providerClient:    providerClient,
			session:           session,
			nodes:             lineBinding.Nodes,
			dialerProxy:       lineBinding.DialerProxy,
			lineLabels:        lineBinding.Labels,
			failure:           failure,
		})
		return err
	})
	return lease, err
}
