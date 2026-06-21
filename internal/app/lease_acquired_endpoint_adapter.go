package app

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
)

type leaseAcquiredEndpointAdapter struct {
	settings              *runtimeSettingsFile
	advertisedHost        string
	leaseListener         leaseListenerFunc
	localListenerEndpoint leaseEndpointFunc
	sessionAdvertisedHost leaseAdvertisedHostFunc
}

func (a leaseAcquiredEndpointAdapter) ResolveListener(ctx context.Context, accountID string, leaseID string) (leaseapp.Listener, error) {
	return a.leaseListener(ctx, a.settings, accountID, leaseID)
}

func (a leaseAcquiredEndpointAdapter) ResolveEgress(ctx context.Context, listener leaseapp.Listener) (*proxygatewayv1.ProxyEndpoint, error) {
	_ = ctx
	return a.localListenerEndpoint(listener, a.sessionAdvertisedHost(a.advertisedHost, listener))
}
