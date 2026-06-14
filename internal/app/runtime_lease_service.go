package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) acquireLease(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	if err := leaseapp.PrepareAcquireRequest(req); err != nil {
		return nil, invalidArgument(err.Error(), err)
	}
	var lease *proxyruntimev1.ProxyDynamicLease
	err := c.deps.locks.WithAccountLock(ctx, req.GetAccountId(), func(ctx context.Context) error {
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
	req.Policy = normalizeDynamicIPSessionPolicy(req.GetPolicy())
	leaseapp.ApplyRequestLabels(req)
	if err := applyLeaseProfileDynamicIPPolicy(settings, req); err != nil {
		return nil, err
	}
	selectionPolicy := normalizeDynamicIPSelectionPolicy(req)
	requestedSessionID := leaseapp.RequestedSessionID(req)
	existing, err := c.activeLeaseByRequest(ctx, req, requestedSessionID)
	if err == nil && leaseapp.ActiveAt(existing, c.now().UTC()) {
		if !req.GetForceNew() && !playgroundLeaseNeedsReplacement(req, existing) {
			if err := c.refreshLeaseConcurrencySlot(ctx, existing); err != nil {
				return nil, err
			}
			return existing, nil
		}
		if err := c.retireLeaseRoute(ctx, existing); err != nil {
			return nil, err
		}
	}
	var lastErr error
	for attempt := 1; attempt <= dynamicIPSelectionMaxAttempts(selectionPolicy); attempt++ {
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
	concurrencySlot, err := c.acquireProviderAccountConcurrencySlot(ctx, providerAccount, dynamicProviderConcurrencyLimit(settings, selection.plan.GetSelectedEndpoint().GetDynamicProviderId(), req.GetPolicy()), req.GetPolicy(), concurrencyHolder, leaseapp.ConcurrencySlotTTL(req.GetPolicy(), defaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
	if err != nil {
		return nil, failedPrecondition("provider account concurrency limit reached", err)
	}
	keepConcurrencySlot := false
	defer func() {
		if !keepConcurrencySlot {
			_ = concurrencySlot.Release(context.Background())
		}
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
	requestedSessionID := leaseapp.RequestedSessionID(req)
	if requestedSessionID != "" {
		req.Policy.Labels[leaseapp.LabelSessionID] = requestedSessionID
	}
	req.Policy.Labels[leaseapp.LabelSelectionID] = selection.plan.GetSelectionId()
	req.Policy.Labels[leaseapp.LabelDynamicIPEndpointID] = selection.plan.GetSelectedEndpoint().GetEndpointId()
	req.Policy.Labels[leaseapp.LabelProviderAccountConcurrencyHolder] = concurrencyHolder
	session, err := providerClient.CreateSession(ctx, req)
	if err != nil {
		return nil, unavailable("provider session create failed", err)
	}
	failure := newLeaseAcquireFailure(c, ctx, req, providerAccountID, providerClient, session, selection.plan)
	nodes, err := providerClient.FetchSession(ctx, session)
	if err != nil {
		failure.beforeRoute("provider session fetch failed")
		return nil, unavailable("provider session fetch failed", err)
	}
	dialerProxy, lineLabels, err := c.deps.dynamicLeaseDialerProxy(ctx, settings, req.GetAccountId())
	if err != nil {
		failure.beforeRoute("lease line resolution failed")
		return nil, err
	}
	nodes = applyDynamicLeaseLineLabels(nodes, lineLabels)
	var lease *proxyruntimev1.ProxyDynamicLease
	err = c.deps.locks.WithSessionListenerAllocationLock(ctx, func(ctx context.Context) error {
		var err error
		lease, err = c.applyAcquiredLeaseRoute(ctx, advertisedHost, req, settings, selection, providerAccountID, leaseID, concurrencyHolder, providerClient, session, nodes, dialerProxy, lineLabels, failure)
		return err
	})
	return lease, err
}
