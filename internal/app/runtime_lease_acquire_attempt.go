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
	providerAccountID := selection.plan.GetSelectedEndpoint().GetProviderAccountId()
	providerAccount, err := c.deps.store.ProviderAccount(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	leaseID, err := c.newLeaseID()
	if err != nil {
		return nil, internalError("generate lease id", err)
	}
	concurrencyHolder := leaseapp.HolderForLeaseID(leaseID)
	concurrencySlot, err := leaseapp.AcquireProviderAccountConcurrencySlot(ctx, c.deps.providerConcurrency, providerAccount.GetAccountId(), dynamicProviderConcurrencyLimit(settings, selection.plan.GetSelectedEndpoint().GetDynamicProviderId(), req.GetPolicy()), req.GetPolicy(), concurrencyHolder, leaseapp.ConcurrencySlotTTL(req.GetPolicy(), defaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
	if err != nil {
		return nil, failedPrecondition("provider account concurrency limit reached", err)
	}
	keepConcurrencySlot := false
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), leaseAcquireSlotReleaseTimeout)
		defer cancel()
		_ = leaseapp.ReleaseConcurrencySlotUnlessKept(releaseCtx, concurrencySlot, keepConcurrencySlot)
	}()
	var lease *proxyruntimev1.ProxyDynamicLease
	err = c.deps.locks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
		var err error
		lease, err = c.acquireLeaseWithProviderAccountLock(ctx, advertisedHost, req, settings, selection, providerAccountID, leaseID, concurrencyHolder)
		if err == nil {
			keepConcurrencySlot = true
		}
		return err
	})
	return lease, err
}
