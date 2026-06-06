package app

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/randx"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c leaseCoordinator) acquireLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	req.AccountId = strings.TrimSpace(req.GetAccountId())
	if req.AccountId == "" {
		return nil, errors.New("account_id is required")
	}
	lock, err := r.leaseLocks.LockAccount(ctx, req.GetAccountId())
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Unlock(ctx) }()
	req.Purpose = firstNonEmpty(req.GetPurpose(), "general")
	normalizeLeasePolicy(req)
	requestedSessionID := requestedLeaseSessionID(req)
	existing, err := r.activeLeaseByRequest(ctx, req, requestedSessionID)
	if err == nil && leaseActive(existing, time.Now().UTC()) {
		if !req.GetForceNew() {
			if err := r.refreshLeaseConcurrencySlot(ctx, existing); err != nil {
				return nil, err
			}
			return existing, nil
		}
		if err := c.retireLeaseRoute(ctx, existing); err != nil {
			return nil, err
		}
	}
	selection, err := r.dynamicIPSelector.selectDynamicIPEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}
	providerAccountID := selection.plan.GetSelectedEndpoint().GetProviderAccountId()
	providerAccount, err := r.store.ProviderAccount(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	leaseID, err := randx.Hex(12)
	if err != nil {
		return nil, err
	}
	concurrencyHolder := "lease:" + leaseID
	concurrencySlot, err := r.acquireProviderAccountConcurrencySlot(ctx, providerAccount, req.GetPolicy(), concurrencyHolder, leaseConcurrencySlotTTL(req.GetPolicy()))
	if err != nil {
		return nil, err
	}
	keepConcurrencySlot := false
	defer func() {
		if !keepConcurrencySlot {
			_ = concurrencySlot.Release(ctx)
		}
	}()
	providerLock, err := r.leaseLocks.LockProviderAccount(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = providerLock.Unlock(ctx) }()
	providerCfg, providerAccountID, err := r.store.ProviderConfig(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	providerCfg.Gateways = []accountproxy.Gateway{selection.endpoint}
	providerClient, err := r.accountProviders.NewSessionProvider(providerCfg, BuildProviderHTTPClient(r.cfg))
	if err != nil {
		return nil, err
	}
	if requestedSessionID != "" {
		req.Policy.Labels["session_id"] = requestedSessionID
	}
	req.Policy.Labels["selection_id"] = selection.plan.GetSelectionId()
	req.Policy.Labels["dynamic_ip_endpoint_id"] = selection.plan.GetSelectedEndpoint().GetEndpointId()
	req.Policy.Labels["provider_account_concurrency_holder"] = concurrencyHolder
	session, err := providerClient.CreateSession(ctx, req)
	if err != nil {
		return nil, err
	}
	failure := newLeaseAcquireFailure(c, ctx, req, providerAccountID, providerClient, session, selection.plan)
	nodes, err := providerClient.FetchSession(ctx, session)
	if err != nil {
		failure.beforeRoute("provider session fetch failed")
		return nil, err
	}
	settings, err := r.settings.load(ctx)
	if err != nil {
		failure.beforeRoute("runtime settings load failed")
		return nil, err
	}
	dialerProxy, lineLabels := dynamicLeaseDialerProxy(settings, selection.plan.GetSelectedEndpoint())
	nodes = applyDynamicLeaseLineLabels(nodes, lineLabels)
	listenerLock, err := r.leaseLocks.LockSessionListenerAllocation(ctx)
	if err != nil {
		failure.beforeRoute("lease listener allocation lock failed")
		return nil, err
	}
	defer func() { _ = listenerLock.Unlock(ctx) }()
	listener, err := r.leaseListener(ctx, req.GetAccountId(), leaseID)
	if err != nil {
		failure.beforeRoute("lease listener allocation failed")
		return nil, err
	}
	listenerProto := protoListener(listener, true)
	failure.listener = listenerProto
	egress, err := r.localListenerEndpoint(listener, r.sessionAdvertisedHost(httpReq, listener))
	if err != nil {
		failure.beforeRoute("lease endpoint build failed")
		return nil, err
	}
	failure.egress = egress
	egress.ProviderId = providerClient.Name()
	egress.UpstreamKind = proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP
	egress.RotationMode = req.GetPolicy().GetRotationMode()
	egress.SessionId = session.GetSessionId()
	egress.Labels["account_id"] = req.GetAccountId()
	egress.Labels["purpose"] = req.GetPurpose()
	egress.Labels["provider_account_id"] = providerAccountID
	egress.Labels["selection_id"] = selection.plan.GetSelectionId()
	egress.Labels["dynamic_provider_id"] = selection.plan.GetSelectedEndpoint().GetDynamicProviderId()
	egress.Labels["dynamic_ip_endpoint_id"] = selection.plan.GetSelectedEndpoint().GetEndpointId()
	egress.Labels["provider_account_concurrency_holder"] = concurrencyHolder
	for key, value := range lineLabels {
		egress.Labels[key] = value
	}
	session.Egress = egress
	route := dataplane.SessionRoute{SessionID: session.GetSessionId(), Listener: localServiceFromListener(listener, r.cfg.LocalProtocol), Pool: nodes, DialerProxy: dialerProxy}
	if err := r.dataPlane.UpsertSessionRoute(ctx, route); err != nil {
		failure.afterRoute(route, "dataplane route apply failed")
		return nil, err
	}
	now := time.Now().UTC()
	lease := &proxyruntimev1.ProxyDynamicLease{LeaseId: leaseID, AccountId: req.GetAccountId(), Purpose: req.GetPurpose(), ProviderAccountId: providerAccountID, Status: proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE, Session: session, Egress: egress, Listener: listenerProto, AcquiredAt: timestamppb.New(now), ExpiresAt: session.GetExpiresAt(), SelectionPlan: selection.plan}
	if err := r.store.SaveLeaseFact(ctx, lease); err != nil {
		failure.afterRoute(route, "lease fact save failed")
		return nil, err
	}
	keepConcurrencySlot = true
	return lease, nil
}

func (r *Runtime) activeLeaseByRequest(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if strings.TrimSpace(sessionID) != "" {
		return r.store.ActiveLeaseFactBySession(ctx, req.GetAccountId(), req.GetPurpose(), sessionID)
	}
	return nil, pgx.ErrNoRows
}

func (c leaseCoordinator) releaseLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	lease, err := c.leaseByReleaseRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if lease.GetStatus() == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED {
		return lease, nil
	}
	accountID := strings.TrimSpace(lease.GetAccountId())
	lock, err := r.leaseLocks.LockAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Unlock(ctx) }()
	current, err := r.store.LeaseFactByID(ctx, lease.GetLeaseId())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if current != nil {
		lease = current
	}
	if lease.GetStatus() == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED {
		return lease, nil
	}
	if lease.GetStatus() != proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
		return lease, nil
	}
	if err := c.retireLeaseRoute(ctx, lease); err != nil {
		return nil, err
	}
	return lease, nil
}

func (c leaseCoordinator) leaseByReleaseRequest(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	if req == nil {
		return nil, errors.New("release request is required")
	}
	leaseID := strings.TrimSpace(req.GetLeaseId())
	accountID := strings.TrimSpace(req.GetAccountId())
	purpose := strings.TrimSpace(req.GetPurpose())
	if leaseID != "" {
		lease, err := r.store.LeaseFactByID(ctx, leaseID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("lease_id not found")
			}
			return nil, err
		}
		if accountID != "" && accountID != lease.GetAccountId() {
			return nil, errors.New("lease account_id mismatch")
		}
		if purpose != "" && purpose != lease.GetPurpose() {
			return nil, errors.New("lease purpose mismatch")
		}
		return lease, nil
	}
	if accountID == "" {
		return nil, errors.New("lease_id or account_id is required")
	}
	lease, err := r.store.ActiveLeaseFactByAccount(ctx, accountID, purpose)
	if err == nil {
		return lease, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	lease, err = r.store.LatestLeaseFactByAccount(ctx, accountID, purpose)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("active lease not found")
		}
		return nil, err
	}
	return lease, nil
}

func (c leaseCoordinator) retireLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	if lease == nil {
		return nil
	}
	if err := c.deleteLeaseRoute(ctx, lease); err != nil {
		_ = c.saveLeaseReleaseCleanupFailure(ctx, lease, true, false, "lease route cleanup failed")
		return err
	}
	releaseErr := c.releaseLeaseProviderSession(ctx, lease)
	if releaseErr != nil {
		r.logger.Warn("provider session release failed", "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		if err := c.saveLeaseReleaseCleanupFailure(ctx, lease, false, true, "provider session release failed"); err != nil {
			return err
		}
		return releaseErr
	}
	return c.saveLeaseReleased(ctx, lease)
}

func (c leaseCoordinator) deleteLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	if lease == nil || lease.GetSession() == nil || lease.GetListener() == nil {
		return nil
	}
	listener := listenerFromProto(lease.GetListener())
	route := dataplane.SessionRoute{SessionID: lease.GetSession().GetSessionId(), Listener: localServiceFromListener(listener, r.cfg.LocalProtocol)}
	return r.dataPlane.DeleteSessionRoute(ctx, route)
}

func normalizeLeasePolicy(req *proxyruntimev1.AcquireProxyLeaseRequest) {
	req.Policy = normalizeDynamicIPSessionPolicy(req.GetPolicy())
	if req.Policy.Labels == nil {
		req.Policy.Labels = map[string]string{}
	}
	req.Policy.Labels["account_id"] = req.GetAccountId()
	req.Policy.Labels["purpose"] = req.GetPurpose()
}

func requestedLeaseSessionID(req *proxyruntimev1.AcquireProxyLeaseRequest) string {
	labels := req.GetPolicy().GetLabels()
	return firstNonEmpty(
		labels["session_id"],
		labels["sticky_session_id"],
		labels["sticky_id"],
		labels["sid"],
		labels["session"],
	)
}

func leaseActive(lease *proxyruntimev1.ProxyDynamicLease, now time.Time) bool {
	if lease == nil || lease.GetStatus() != proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
		return false
	}
	return lease.GetExpiresAt() == nil || now.Before(lease.GetExpiresAt().AsTime())
}
