package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type acquiredLeaseEndpointInput struct {
	advertisedHost    string
	req               *proxyruntimev1.AcquireProxyLeaseRequest
	settings          *runtimeSettingsFile
	selection         dynamicIPSelection
	providerAccountID string
	leaseID           string
	concurrencyHolder string
	providerClient    provider.SessionProvider
	session           *proxyruntimev1.ProxySession
	lineLabels        map[string]string
}

func (c leaseCoordinator) acquiredLeaseEndpoint(ctx context.Context, input acquiredLeaseEndpointInput, failure *leaseAcquireFailure) (leaseapp.Listener, *proxyruntimev1.EgressListener, *proxyruntimev1.ProxyEndpoint, error) {
	listener, err := c.deps.leaseListener(ctx, input.settings, input.req.GetAccountId(), input.leaseID)
	if err != nil {
		failure.beforeRoute("lease listener allocation failed")
		return leaseapp.Listener{}, nil, nil, err
	}
	listenerProto := protoLeaseListener(listener, true)
	failure.listener = listenerProto
	egress, err := c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(input.advertisedHost, listener))
	if err != nil {
		failure.beforeRoute("lease endpoint build failed")
		return leaseapp.Listener{}, nil, nil, err
	}
	failure.egress = egress
	applyAcquiredLeaseEndpointMetadata(egress, input)
	return listener, listenerProto, egress, nil
}

func applyAcquiredLeaseEndpointMetadata(egress *proxyruntimev1.ProxyEndpoint, input acquiredLeaseEndpointInput) {
	if egress.Labels == nil {
		egress.Labels = map[string]string{}
	}
	egress.ProviderId = input.providerClient.Name()
	egress.UpstreamKind = proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP
	egress.RotationMode = input.req.GetPolicy().GetRotationMode()
	egress.SessionId = input.session.GetSessionId()
	egress.Labels[leaseapp.LabelAccountID] = input.req.GetAccountId()
	egress.Labels[leaseapp.LabelPurpose] = input.req.GetPurpose()
	egress.Labels[leaseapp.LabelProviderAccountID] = input.providerAccountID
	egress.Labels[leaseapp.LabelSelectionID] = input.selection.plan.GetSelectionId()
	egress.Labels[leaseapp.LabelDynamicProviderID] = input.selection.plan.GetSelectedEndpoint().GetDynamicProviderId()
	egress.Labels[leaseapp.LabelDynamicIPEndpointID] = input.selection.plan.GetSelectedEndpoint().GetEndpointId()
	egress.Labels[leaseapp.LabelProviderAccountConcurrencyHolder] = input.concurrencyHolder
	if countryCode := strings.TrimSpace(input.selection.plan.GetPolicy().GetCountryCode()); countryCode != "" {
		egress.Labels["country_code"] = countryCode
	}
	if region := strings.TrimSpace(firstNonEmpty(input.req.GetPolicy().GetRegion(), input.selection.plan.GetPolicy().GetRegion())); region != "" {
		egress.Labels["region"] = region
	}
	if state := strings.TrimSpace(input.req.GetPolicy().GetState()); state != "" {
		egress.Labels["state"] = state
	}
	if city := strings.TrimSpace(input.req.GetPolicy().GetCity()); city != "" {
		egress.Labels["city"] = city
	}
	if asn := strings.TrimSpace(input.req.GetPolicy().GetAsn()); asn != "" {
		egress.Labels["asn"] = asn
	}
	for key, value := range input.lineLabels {
		egress.Labels[key] = value
	}
}
