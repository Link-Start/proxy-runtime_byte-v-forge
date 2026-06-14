package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireLeaseWithProviderAccountLock(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string) (*proxyruntimev1.ProxyDynamicLease, error) {
	applier := c.acquiredRouteApplier(settings, advertisedHost, req)
	lease, err := leaseapp.RunProviderAccountAcquire(ctx, leaseapp.ProviderAccountAcquireInput{
		Store:              c.deps.store,
		IDs:                c.deps.ids,
		Clock:              c.deps.clock,
		DataPlane:          c.deps.dataPlane,
		Logger:             c.deps.logger,
		Factory:            c.deps.sessionProviders,
		Locks:              c.deps.locks,
		ProviderAccountID:  providerAccountID,
		Gateway:            selection.endpoint,
		Request:            req,
		SelectionPlan:      selection.plan,
		ConcurrencyHolder:  concurrencyHolder,
		ResolveLineBinding: c.routeLineBindingResolver(settings),
		Apply: func(ctx context.Context, acquired leaseapp.ProviderAccountAcquireApplyInput) (*proxyruntimev1.ProxyDynamicLease, error) {
			lease, err := applier.Apply(ctx, leaseapp.AcquiredRouteApplierInput{
				Failure:           acquired.Failure,
				LeaseID:           leaseID,
				Request:           req,
				ProviderClient:    acquired.ProviderClient,
				ProviderAccountID: acquired.ProviderAccountID,
				ConcurrencyHolder: concurrencyHolder,
				Session:           acquired.Session,
				Nodes:             acquired.Nodes,
				DialerProxy:       acquired.DialerProxy,
				LineLabels:        acquired.LineLabels,
				SelectionPlan:     selection.plan,
			})
			if err != nil {
				return nil, acquiredRouteApplyError(err)
			}
			return lease, nil
		},
	})
	if err != nil {
		return nil, providerSessionAcquireError(err)
	}
	return lease, nil
}
