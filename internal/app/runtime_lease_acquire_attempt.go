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
	providerAccountID := leaseapp.SelectedProviderAccountID(selection.plan)
	leaseID, err := c.newLeaseID()
	if err != nil {
		return nil, internalError("generate lease id", err)
	}
	attemptSlot, err := leaseapp.AcquireAttemptSlotForProviderAccount(ctx, leaseapp.AcquireAttemptSlotInput{
		Store:             c.deps.store,
		Limiter:           c.deps.providerConcurrency,
		ProviderAccountID: providerAccountID,
		Limit:             dynamicProviderConcurrencyLimit(settings, leaseapp.SelectedDynamicProviderID(selection.plan), req.GetPolicy()),
		Policy:            req.GetPolicy(),
		LeaseID:           leaseID,
		DefaultTTL:        leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:         providerAccountConcurrencyTTLBuffer,
	})
	if err != nil {
		return nil, acquireAttemptSlotError(err)
	}
	return leaseapp.RunLockedAcquireAttempt(ctx, leaseapp.LockedAcquireAttemptInput{
		Locks:             c.deps.locks,
		ProviderAccountID: providerAccountID,
		ConcurrencySlot:   attemptSlot.ConcurrencySlot,
		ReleaseTimeout:    leaseAcquireSlotReleaseTimeout,
		Action: func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
			return c.acquireLeaseWithProviderAccountLock(ctx, advertisedHost, req, settings, selection, providerAccountID, leaseID, attemptSlot.ConcurrencyHolder)
		},
	})
}
