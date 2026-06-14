package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

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
