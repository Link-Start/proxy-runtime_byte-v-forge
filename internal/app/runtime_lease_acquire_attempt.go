package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

func (c leaseCoordinator) acquireLeaseAttempt(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile) (*proxyruntimev1.ProxyDynamicLease, error) {
	selection, err := c.deps.dynamicIPSelector.selectDynamicIPEndpoint(ctx, req)
	if err != nil {
		return nil, failedPrecondition("no dynamic IP endpoint candidate", err)
	}
	lease, err := leaseapp.RunSelectedAcquireAttempt(ctx, leaseapp.SelectedAcquireAttemptRunInput{
		Store:          c.deps.store,
		IDs:            c.deps.ids,
		Limiter:        c.deps.providerConcurrency,
		Locks:          c.deps.locks,
		SelectionPlan:  selection.plan,
		Limit:          dynamicProviderConcurrencyLimit(settings, leaseapp.SelectedDynamicProviderID(selection.plan), req.GetPolicy()),
		Policy:         req.GetPolicy(),
		DefaultTTL:     leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:      providerAccountConcurrencyTTLBuffer,
		ReleaseTimeout: leaseAcquireSlotReleaseTimeout,
		Action: func(ctx context.Context, attempt leaseapp.SelectedAcquireAttempt) (*proxyruntimev1.ProxyDynamicLease, error) {
			return c.acquireLeaseWithProviderAccountLock(ctx, advertisedHost, req, settings, selection, attempt.ProviderAccountID, attempt.LeaseID, attempt.ConcurrencyHolder)
		},
	})
	if err != nil {
		return nil, acquireAttemptSlotError(err)
	}
	return lease, nil
}
