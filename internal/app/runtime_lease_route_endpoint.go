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
	endpoint, err := leaseapp.MaterializeAcquiredEndpoint(ctx, leaseapp.AcquiredEndpointMaterializeInput{
		Request:           input.req,
		LeaseID:           input.leaseID,
		SelectionPlan:     input.selection.plan,
		ProviderClient:    input.providerClient,
		ProviderAccountID: input.providerAccountID,
		ConcurrencyHolder: input.concurrencyHolder,
		Session:           input.session,
		LineLabels:        input.lineLabels,
		Failure:           failure,
		Managed:           true,
		FallbackProtocol:  "http",
		ResolveListener: func(ctx context.Context, accountID string, leaseID string) (leaseapp.Listener, error) {
			return c.deps.leaseListener(ctx, input.settings, accountID, leaseID)
		},
		ResolveEgress: func(ctx context.Context, listener leaseapp.Listener) (*proxyruntimev1.ProxyEndpoint, error) {
			_ = ctx
			return c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(input.advertisedHost, listener))
		},
	})
	if err != nil {
		return leaseapp.Listener{}, nil, nil, err
	}
	return endpoint.Listener, endpoint.ListenerProto, endpoint.Egress, nil
}
