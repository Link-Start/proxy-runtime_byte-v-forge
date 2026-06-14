package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireLeaseWithProviderAccountLock(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := leaseapp.RunProviderAccountAcquire(ctx, leaseapp.ProviderAccountAcquireInput{
		Store:             c.deps.store,
		IDs:               c.deps.ids,
		Clock:             c.deps.clock,
		DataPlane:         c.deps.dataPlane,
		Logger:            c.deps.logger,
		Factory:           c.deps.sessionProviders,
		Locks:             c.deps.locks,
		ProviderAccountID: providerAccountID,
		Gateway:           selection.endpoint,
		Request:           req,
		SelectionPlan:     selection.plan,
		ConcurrencyHolder: concurrencyHolder,
		ResolveLineBinding: func(ctx context.Context, accountID string) (string, map[string]string, error) {
			return c.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
		},
		Apply: func(ctx context.Context, acquired leaseapp.ProviderAccountAcquireApplyInput) (*proxyruntimev1.ProxyDynamicLease, error) {
			return c.applyAcquiredLeaseRoute(ctx, acquiredLeaseFlow{
				advertisedHost:    advertisedHost,
				request:           req,
				settings:          settings,
				selection:         selection,
				providerAccountID: acquired.ProviderAccountID,
				leaseID:           leaseID,
				concurrencyHolder: concurrencyHolder,
				providerClient:    acquired.ProviderClient,
				session:           acquired.Session,
				nodes:             acquired.Nodes,
				dialerProxy:       acquired.DialerProxy,
				lineLabels:        acquired.LineLabels,
				failure:           acquired.Failure,
			})
		},
	})
	if err != nil {
		return nil, providerSessionAcquireError(err)
	}
	return lease, nil
}
