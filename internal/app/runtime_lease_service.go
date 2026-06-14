package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
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

func (c leaseCoordinator) applyAcquiredLeaseRoute(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string, providerClient provider.SessionProvider, session *proxyruntimev1.ProxySession, nodes []provider.Node, dialerProxy string, lineLabels map[string]string, failure *leaseAcquireFailure) (*proxyruntimev1.ProxyDynamicLease, error) {
	listener, err := c.deps.leaseListener(ctx, settings, req.GetAccountId(), leaseID)
	if err != nil {
		failure.beforeRoute("lease listener allocation failed")
		return nil, err
	}
	listenerProto := protoListener(listener, true)
	failure.listener = listenerProto
	egress, err := c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(advertisedHost, listener))
	if err != nil {
		failure.beforeRoute("lease endpoint build failed")
		return nil, err
	}
	failure.egress = egress
	egress.ProviderId = providerClient.Name()
	egress.UpstreamKind = proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP
	egress.RotationMode = req.GetPolicy().GetRotationMode()
	egress.SessionId = session.GetSessionId()
	egress.Labels[leaseapp.LabelAccountID] = req.GetAccountId()
	egress.Labels[leaseapp.LabelPurpose] = req.GetPurpose()
	egress.Labels[leaseapp.LabelProviderAccountID] = providerAccountID
	egress.Labels[leaseapp.LabelSelectionID] = selection.plan.GetSelectionId()
	egress.Labels[leaseapp.LabelDynamicProviderID] = selection.plan.GetSelectedEndpoint().GetDynamicProviderId()
	egress.Labels[leaseapp.LabelDynamicIPEndpointID] = selection.plan.GetSelectedEndpoint().GetEndpointId()
	egress.Labels[leaseapp.LabelProviderAccountConcurrencyHolder] = concurrencyHolder
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
	route := dataplane.SessionRoute{SessionID: session.GetSessionId(), Listener: localServiceFromListener(listener, c.deps.cfg.LocalProtocol), Pool: nodes, DialerProxy: dialerProxy}
	if err := c.deps.dataPlane.UpsertSessionRoute(ctx, route); err != nil {
		failure.afterRoute(route, "dataplane route apply failed")
		return nil, unavailable("dataplane route apply failed", err)
	}
	lease := leaseapp.NewActiveFact(leaseapp.ActiveFactInput{
		LeaseID:           leaseID,
		AccountID:         req.GetAccountId(),
		Purpose:           req.GetPurpose(),
		ProviderAccountID: providerAccountID,
		Session:           session,
		Egress:            egress,
		Listener:          listenerProto,
		SelectionPlan:     selection.plan,
		AcquiredAt:        c.now(),
	})
	if err := c.deps.store.SaveLeaseFact(ctx, lease); err != nil {
		failure.afterRoute(route, "lease fact save failed")
		return nil, internalError("lease fact save failed", err)
	}
	c.clearExitCheckCache()
	if req.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
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

func playgroundLeaseNeedsReplacement(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease) bool {
	if req.GetAccountId() != playgroundProfileID {
		return false
	}
	return strings.TrimSpace(lease.GetListener().GetLabels()["proxy_username"]) != playgroundUsername
}
