package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type acquiredLeaseEndpointInput struct {
	advertisedHost    string
	req               *proxyruntimev1.AcquireProxyLeaseRequest
	settings          *runtimeSettingsFile
	selection         dynamicIPSelection
	providerAccountID string
	leaseID           string
	concurrencyHolder string
	providerClient    leaseapp.SessionProvider
	session           *proxyruntimev1.ProxySession
	lineLabels        map[string]string
}

func (c leaseCoordinator) acquiredLeaseEndpoint(ctx context.Context, input acquiredLeaseEndpointInput, failure *leaseapp.FailedAcquireRecorder) (leaseapp.Listener, *proxyruntimev1.EgressListener, *proxyruntimev1.ProxyEndpoint, error) {
	listener, err := c.deps.leaseListener(ctx, input.settings, input.req.GetAccountId(), input.leaseID)
	if err != nil {
		failure.BeforeRoute(ctx, "lease listener allocation failed")
		return leaseapp.Listener{}, nil, nil, err
	}
	listenerProto := leaseapp.EgressListenerProto(listener, true, "http")
	failure.SetListener(listenerProto)
	egress, err := c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(input.advertisedHost, listener))
	if err != nil {
		failure.BeforeRoute(ctx, "lease endpoint build failed")
		return leaseapp.Listener{}, nil, nil, err
	}
	failure.SetEgress(egress)
	applyAcquiredLeaseEndpointMetadata(egress, input)
	return listener, listenerProto, egress, nil
}

func applyAcquiredLeaseEndpointMetadata(egress *proxyruntimev1.ProxyEndpoint, input acquiredLeaseEndpointInput) {
	leaseapp.ApplyDynamicEndpointMetadata(egress, leaseapp.DynamicEndpointMetadataInput{
		Request:           input.req,
		SelectionPlan:     input.selection.plan,
		ProviderID:        input.providerClient.Name(),
		ProviderAccountID: input.providerAccountID,
		ConcurrencyHolder: input.concurrencyHolder,
		SessionID:         input.session.GetSessionId(),
		LineLabels:        input.lineLabels,
	})
}
