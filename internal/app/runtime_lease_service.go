package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireLease(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	if err := leaseapp.PrepareAcquireRequest(req); err != nil {
		return nil, invalidArgument(err.Error(), err)
	}
	var lease *proxyruntimev1.ProxyDynamicLease
	err := leaseapp.WithAccountLock(ctx, c.deps.locks, req.GetAccountId(), func(ctx context.Context) error {
		var err error
		lease, err = c.acquireLeaseWithAccountLock(ctx, advertisedHost, req)
		return err
	})
	return lease, err
}

func (c leaseCoordinator) acquireLeaseWithAccountLock(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	selectionPolicy, err := leaseapp.ApplyAcquireRequestPolicies(req, settings.GetEgressProfiles())
	if err != nil {
		return nil, leaseProfilePolicyError(err)
	}
	requestedSessionID := leaseapp.RequestedSessionID(req)
	existing, err := leaseapp.ActiveLeaseByRequest(ctx, c.deps.store, req, requestedSessionID)
	if err == nil {
		switch leaseapp.DecideExistingActiveLease(req, existing, c.now().UTC(), playgroundProfileID, playgroundUsername) {
		case leaseapp.ExistingActiveLeaseReuse:
			if err := c.refreshLeaseConcurrencySlot(ctx, existing); err != nil {
				return nil, err
			}
			return existing, nil
		case leaseapp.ExistingActiveLeaseReplace:
			if err := c.retireLeaseRoute(ctx, existing); err != nil {
				return nil, err
			}
		}
	}
	return leaseapp.RunAcquireAttempts(
		req,
		selectionPolicy,
		func(int) (*proxyruntimev1.ProxyDynamicLease, error) {
			return c.acquireLeaseAttempt(ctx, advertisedHost, req, settings)
		},
		retryLeaseAcquireAttempt,
		func(attempt int, err error) {
			c.warn("dynamic IP lease attempt failed", leaseapp.LabelAccountID, req.GetAccountId(), leaseapp.LabelPurpose, req.GetPurpose(), "attempt", attempt, "error_type", errorLogType(err))
		},
	)
}
