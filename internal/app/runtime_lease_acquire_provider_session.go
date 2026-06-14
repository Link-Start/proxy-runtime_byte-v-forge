package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) acquireLeaseWithProviderAccountLock(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string) (*proxyruntimev1.ProxyDynamicLease, error) {
	providerCfg, providerAccountID, err := c.deps.store.ProviderConfig(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	providerCfg.Gateways = []accountproxy.Gateway{selection.endpoint}
	providerClient, err := c.newSessionProvider(providerCfg)
	if err != nil {
		return nil, invalidArgument("provider account configuration is invalid", err)
	}
	session, err := leaseapp.CreateProviderSession(ctx, providerClient, req, selection.plan, concurrencyHolder)
	if err != nil {
		return nil, unavailable("provider session create failed", err)
	}
	failure := newLeaseAcquireFailure(c, ctx, req, providerAccountID, providerClient, session, selection.plan)
	nodes, err := leaseapp.FetchProviderSession(ctx, providerClient, session)
	if err != nil {
		failure.beforeRoute("provider session fetch failed")
		return nil, unavailable("provider session fetch failed", err)
	}
	dialerProxy, lineLabels, err := c.deps.dynamicLeaseDialerProxy(ctx, settings, req.GetAccountId())
	if err != nil {
		failure.beforeRoute("lease line resolution failed")
		return nil, err
	}
	nodes = leaseapp.ApplyNodeLabels(nodes, lineLabels)
	var lease *proxyruntimev1.ProxyDynamicLease
	err = c.deps.locks.WithSessionListenerAllocationLock(ctx, func(ctx context.Context) error {
		var err error
		lease, err = c.applyAcquiredLeaseRoute(ctx, advertisedHost, req, settings, selection, providerAccountID, leaseID, concurrencyHolder, providerClient, session, nodes, dialerProxy, lineLabels, failure)
		return err
	})
	return lease, err
}
