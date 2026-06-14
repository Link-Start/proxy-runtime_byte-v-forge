package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) restoreLeaseSessionNodes(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, settings *runtimeSettingsFile, providerCfg accountproxy.Config) ([]provider.Node, error) {
	providerCfg.Gateways = endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerCfg.ProviderID)
	providerClient, err := leaseapp.NewSessionProvider(c.deps.sessionProviders, providerCfg)
	if err != nil {
		return nil, err
	}
	return leaseapp.FetchProviderSession(ctx, providerClient, lease.GetSession())
}
