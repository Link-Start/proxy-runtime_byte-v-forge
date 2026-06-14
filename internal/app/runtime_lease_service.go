package app

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/random"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c leaseCoordinator) acquireLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	if req == nil {
		return nil, invalidArgument("acquire request is required", nil)
	}
	req.AccountId = strings.TrimSpace(req.GetAccountId())
	if req.AccountId == "" {
		return nil, invalidArgument("account_id is required", nil)
	}
	var lease *proxyruntimev1.ProxyDynamicLease
	err := r.leaseLocks.WithAccountLock(ctx, req.GetAccountId(), func(ctx context.Context) error {
		var err error
		lease, err = c.acquireLeaseWithAccountLock(ctx, httpReq, req)
		return err
	})
	return lease, err
}

func (c leaseCoordinator) acquireLeaseWithAccountLock(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	req.Purpose = firstNonEmpty(req.GetPurpose(), "general")
	settings, err := r.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	normalizeLeasePolicy(req)
	if err := applyLeaseProfileDynamicIPPolicy(settings, req); err != nil {
		return nil, err
	}
	selectionPolicy := normalizeDynamicIPSelectionPolicy(req)
	requestedSessionID := requestedLeaseSessionID(req)
	existing, err := r.activeLeaseByRequest(ctx, req, requestedSessionID)
	if err == nil && leaseapp.ActiveAt(existing, time.Now().UTC()) {
		if !req.GetForceNew() && !playgroundLeaseNeedsReplacement(req, existing) {
			if err := r.refreshLeaseConcurrencySlot(ctx, existing); err != nil {
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
		req.Policy.Labels["attempt"] = strconv.Itoa(attempt)
		lease, err := c.acquireLeaseAttempt(ctx, httpReq, req, settings)
		if err == nil {
			return lease, nil
		}
		lastErr = err
		if !retryLeaseAcquireAttempt(err) {
			return nil, err
		}
		r.logger.Warn("dynamic IP lease attempt failed", "account_id", req.GetAccountId(), "purpose", req.GetPurpose(), "attempt", attempt, "reason", err.Error())
	}
	return nil, lastErr
}

func (c leaseCoordinator) acquireLeaseAttempt(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	selection, err := r.dynamicIPSelector.selectDynamicIPEndpoint(ctx, req)
	if err != nil {
		return nil, failedPrecondition("no dynamic IP endpoint candidate", err)
	}
	providerAccountID := selection.plan.GetSelectedEndpoint().GetProviderAccountId()
	providerAccount, err := r.store.ProviderAccount(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	leaseID, err := random.Hex(12)
	if err != nil {
		return nil, internalError("generate lease id", err)
	}
	concurrencyHolder := "lease:" + leaseID
	concurrencySlot, err := r.acquireProviderAccountConcurrencySlot(ctx, providerAccount, dynamicProviderConcurrencyLimit(settings, selection.plan.GetSelectedEndpoint().GetDynamicProviderId(), req.GetPolicy()), req.GetPolicy(), concurrencyHolder, leaseConcurrencySlotTTL(req.GetPolicy()))
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
	err = r.leaseLocks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
		var err error
		lease, err = c.acquireLeaseWithProviderAccountLock(ctx, httpReq, req, settings, selection, providerAccountID, leaseID, concurrencyHolder)
		if err == nil {
			keepConcurrencySlot = true
		}
		return err
	})
	return lease, err
}

func (c leaseCoordinator) acquireLeaseWithProviderAccountLock(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	providerCfg, providerAccountID, err := r.store.ProviderConfig(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	providerCfg.Gateways = []accountproxy.Gateway{selection.endpoint}
	providerClient, err := r.accountProviders.NewSessionProvider(providerCfg, r.providerHTTPClient)
	if err != nil {
		return nil, invalidArgument("provider account configuration is invalid", err)
	}
	requestedSessionID := requestedLeaseSessionID(req)
	if requestedSessionID != "" {
		req.Policy.Labels["session_id"] = requestedSessionID
	}
	req.Policy.Labels["selection_id"] = selection.plan.GetSelectionId()
	req.Policy.Labels["dynamic_ip_endpoint_id"] = selection.plan.GetSelectedEndpoint().GetEndpointId()
	req.Policy.Labels["provider_account_concurrency_holder"] = concurrencyHolder
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
	dialerProxy, lineLabels, err := r.dynamicLeaseDialerProxy(ctx, settings, req.GetAccountId())
	if err != nil {
		failure.beforeRoute("lease line resolution failed")
		return nil, err
	}
	nodes = applyDynamicLeaseLineLabels(nodes, lineLabels)
	var lease *proxyruntimev1.ProxyDynamicLease
	err = r.leaseLocks.WithSessionListenerAllocationLock(ctx, func(ctx context.Context) error {
		var err error
		lease, err = c.applyAcquiredLeaseRoute(ctx, httpReq, req, settings, selection, providerAccountID, leaseID, concurrencyHolder, providerClient, session, nodes, dialerProxy, lineLabels, failure)
		return err
	})
	return lease, err
}

func (c leaseCoordinator) applyAcquiredLeaseRoute(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string, providerClient provider.SessionProvider, session *proxyruntimev1.ProxySession, nodes []provider.Node, dialerProxy string, lineLabels map[string]string, failure *leaseAcquireFailure) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	listener, err := r.leaseListener(ctx, settings, req.GetAccountId(), leaseID)
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
	if countryCode := strings.TrimSpace(selection.plan.GetPolicy().GetCountryCode()); countryCode != "" {
		egress.Labels["country_code"] = countryCode
	}
	if region := strings.TrimSpace(firstNonEmpty(req.GetPolicy().GetRegion(), selection.plan.GetPolicy().GetRegion())); region != "" {
		egress.Labels["region"] = region
	}
	if state := strings.TrimSpace(req.GetPolicy().GetState()); state != "" {
		egress.Labels["state"] = state
	}
	if city := strings.TrimSpace(req.GetPolicy().GetCity()); city != "" {
		egress.Labels["city"] = city
	}
	if asn := strings.TrimSpace(req.GetPolicy().GetAsn()); asn != "" {
		egress.Labels["asn"] = asn
	}
	for key, value := range lineLabels {
		egress.Labels[key] = value
	}
	session.Egress = egress
	route := dataplane.SessionRoute{SessionID: session.GetSessionId(), Listener: localServiceFromListener(listener, r.cfg.LocalProtocol), Pool: nodes, DialerProxy: dialerProxy}
	if err := r.dataPlane.UpsertSessionRoute(ctx, route); err != nil {
		failure.afterRoute(route, "dataplane route apply failed")
		return nil, unavailable("dataplane route apply failed", err)
	}
	now := time.Now().UTC()
	lease := &proxyruntimev1.ProxyDynamicLease{LeaseId: leaseID, AccountId: req.GetAccountId(), Purpose: req.GetPurpose(), ProviderAccountId: providerAccountID, Status: proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE, Session: session, Egress: egress, Listener: listenerProto, AcquiredAt: timestamppb.New(now), ExpiresAt: session.GetExpiresAt(), SelectionPlan: selection.plan}
	if err := r.store.SaveLeaseFact(ctx, lease); err != nil {
		failure.afterRoute(route, "lease fact save failed")
		return nil, internalError("lease fact save failed", err)
	}
	r.exitCheckCache.clear()
	if req.GetAccountId() == playgroundProfileID {
		r.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	return lease, nil
}

func applyLeaseProfileDynamicIPPolicy(settings *runtimeSettingsFile, req *proxyruntimev1.AcquireProxyLeaseRequest) error {
	profile := egressProfileByID(settings, req.GetAccountId())
	if profile == nil {
		if req.GetPurpose() == "in-user-profile" {
			return failedPrecondition("in-user profile dynamic IP is not configured", nil)
		}
		return nil
	}
	if !profile.GetEnabled() || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return failedPrecondition("in-user profile dynamic IP is not configured", nil)
	}
	if profile.GetExit().GetDynamicIpPolicy().GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return failedPrecondition("in-user profile lease requires sticky dynamic IP", nil)
	}
	if req.GetPolicy().GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return invalidArgument("lease request must use sticky dynamic IP", nil)
	}
	req.Policy = profileDynamicIPLeasePolicy(profile.GetExit().GetDynamicIpPolicy(), req.GetPolicy())
	return nil
}

func profileDynamicIPLeasePolicy(profilePolicy *proxyruntimev1.ProxySessionPolicy, requestPolicy *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	policy := normalizeDynamicIPSessionPolicy(profilePolicy)
	request := normalizeDynamicIPSessionPolicy(requestPolicy)
	policy.StickyTtl = cloneDuration(request.GetStickyTtl())
	policy.Labels = cloneStringMap(policy.GetLabels())
	if policy.Labels == nil {
		policy.Labels = map[string]string{}
	}
	for key, value := range request.GetLabels() {
		policy.Labels[key] = value
	}
	return policy
}

func egressProfileByID(settings *runtimeSettingsFile, profileID string) *proxyruntimev1.EgressProfileSettings {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range settings.GetEgressProfiles() {
		if profile.GetProfileId() == profileID {
			return profile
		}
	}
	return nil
}

func (r *Runtime) activeLeaseByRequest(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if strings.TrimSpace(sessionID) != "" {
		return r.store.ActiveLeaseFactBySession(ctx, req.GetAccountId(), req.GetPurpose(), sessionID)
	}
	return r.store.ActiveLeaseFactByAccount(ctx, req.GetAccountId(), req.GetPurpose())
}

func playgroundLeaseNeedsReplacement(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease) bool {
	if req.GetAccountId() != playgroundProfileID {
		return false
	}
	return strings.TrimSpace(lease.GetListener().GetLabels()["proxy_username"]) != playgroundUsername
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
	err = r.leaseLocks.WithAccountLock(ctx, accountID, func(ctx context.Context) error {
		current, err := r.store.LeaseFactByID(ctx, lease.GetLeaseId())
		if err != nil && !isStoreNotFound(err) {
			return err
		}
		if current != nil {
			lease = current
		}
		if lease.GetStatus() == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED {
			return nil
		}
		if lease.GetStatus() != proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
			return nil
		}
		return c.retireLeaseRoute(ctx, lease)
	})
	return lease, err
}

func (c leaseCoordinator) leaseByReleaseRequest(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	r := c.runtime
	if req == nil {
		return nil, invalidArgument("release request is required", nil)
	}
	leaseID := strings.TrimSpace(req.GetLeaseId())
	accountID := strings.TrimSpace(req.GetAccountId())
	purpose := strings.TrimSpace(req.GetPurpose())
	if leaseID != "" {
		lease, err := r.store.LeaseFactByID(ctx, leaseID)
		if err != nil {
			if isStoreNotFound(err) {
				return nil, invalidArgument("lease_id not found", nil)
			}
			return nil, err
		}
		if accountID != "" && accountID != lease.GetAccountId() {
			return nil, invalidArgument("lease account_id mismatch", nil)
		}
		if purpose != "" && purpose != lease.GetPurpose() {
			return nil, invalidArgument("lease purpose mismatch", nil)
		}
		return lease, nil
	}
	if accountID == "" {
		return nil, invalidArgument("lease_id or account_id is required", nil)
	}
	lease, err := r.store.ActiveLeaseFactByAccount(ctx, accountID, purpose)
	if err == nil {
		return lease, nil
	}
	if !isStoreNotFound(err) {
		return nil, err
	}
	lease, err = r.store.LatestLeaseFactByAccount(ctx, accountID, purpose)
	if err != nil {
		if isStoreNotFound(err) {
			return nil, invalidArgument("active lease not found", nil)
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
	r.exitCheckCache.clear()
	if lease.GetAccountId() == playgroundProfileID {
		r.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	releaseErr := c.releaseLeaseProviderSessionWithLock(ctx, lease)
	if releaseErr != nil {
		r.logger.Warn("provider session release failed", "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		if err := c.saveLeaseReleaseCleanupFailure(ctx, lease, false, true, "provider session release failed"); err != nil {
			return err
		}
		return releaseErr
	}
	return c.saveLeaseReleased(ctx, lease)
}

func (c leaseCoordinator) releaseLeaseProviderSessionWithLock(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	providerAccountID := strings.TrimSpace(lease.GetProviderAccountId())
	if providerAccountID == "" {
		return c.releaseLeaseProviderSession(ctx, lease)
	}
	return c.runtime.leaseLocks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
		return c.releaseLeaseProviderSession(ctx, lease)
	})
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
