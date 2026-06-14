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
	req.Policy = leaseapp.NormalizeDynamicIPSessionPolicy(req.GetPolicy())
	leaseapp.ApplyRequestLabels(req)
	if err := leaseapp.ApplyProfileDynamicIPPolicy(settings.GetEgressProfiles(), req); err != nil {
		return nil, leaseProfilePolicyError(err)
	}
	selectionPolicy := leaseapp.NormalizeDynamicIPSelectionPolicy(req)
	requestedSessionID := leaseapp.RequestedSessionID(req)
	existing, err := c.activeLeaseByRequest(ctx, req, requestedSessionID)
	if err == nil {
		now := c.now().UTC()
		if leaseapp.ReuseExistingActiveLease(req, existing, now, playgroundProfileID, playgroundUsername) {
			if err := c.refreshLeaseConcurrencySlot(ctx, existing); err != nil {
				return nil, err
			}
			return existing, nil
		}
		if leaseapp.ReplaceExistingActiveLease(req, existing, now, playgroundProfileID, playgroundUsername) {
			if err := c.retireLeaseRoute(ctx, existing); err != nil {
				return nil, err
			}
		}
	}
	var lastErr error
	for attempt := 1; attempt <= leaseapp.DynamicIPSelectionMaxAttempts(selectionPolicy); attempt++ {
		leaseapp.SetAttemptLabel(req, attempt)
		lease, err := c.acquireLeaseAttempt(ctx, advertisedHost, req, settings)
		if err == nil {
			return lease, nil
		}
		lastErr = err
		if !retryLeaseAcquireAttempt(err) {
			return nil, err
		}
		c.warn("dynamic IP lease attempt failed", leaseapp.LabelAccountID, req.GetAccountId(), leaseapp.LabelPurpose, req.GetPurpose(), "attempt", attempt, "error_type", errorLogType(err))
	}
	return nil, lastErr
}
