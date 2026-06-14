package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) releaseLeaseProviderSession(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.ReleaseLeaseProviderSession(ctx, leaseapp.ProviderSessionReleaseInput{
		Store:   c.deps.store,
		Factory: c.deps.sessionProviders,
		Lease:   lease,
		ResolveGateways: func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
			settings, err := c.deps.settings.load(ctx)
			if err != nil {
				return nil, err
			}
			return endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerID), nil
		},
	})
}
