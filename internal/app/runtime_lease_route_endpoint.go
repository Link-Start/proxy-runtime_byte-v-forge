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
	egress, err := c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(input.advertisedHost, listener))
	if err != nil {
		failure.BeforeRoute(ctx, "lease endpoint build failed")
		return leaseapp.Listener{}, nil, nil, err
	}
	endpoint := leaseapp.NewAcquiredEndpoint(leaseapp.AcquiredEndpointInput{
		Listener:          listener,
		Egress:            egress,
		Request:           input.req,
		SelectionPlan:     input.selection.plan,
		ProviderClient:    input.providerClient,
		ProviderAccountID: input.providerAccountID,
		ConcurrencyHolder: input.concurrencyHolder,
		Session:           input.session,
		LineLabels:        input.lineLabels,
		Managed:           true,
		FallbackProtocol:  "http",
	})
	failure.SetEndpoint(endpoint)
	return endpoint.Listener, endpoint.ListenerProto, endpoint.Egress, nil
}
