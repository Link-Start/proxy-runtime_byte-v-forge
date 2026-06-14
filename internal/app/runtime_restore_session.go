package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) restoreLeaseSessionNodes(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, settings *runtimeSettingsFile, providerCfg accountproxy.Config) ([]provider.Node, error) {
	return leaseapp.FetchLeaseProviderSession(ctx, leaseapp.ProviderSessionFetchInput{
		Factory:        c.deps.sessionProviders,
		Lease:          lease,
		ProviderConfig: providerCfg,
		ResolveGateways: func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
			return endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerID), nil
		},
	})
}
