package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type ProviderSessionFetchInput struct {
	Factory         SessionProviderFactory
	Lease           *proxyruntimev1.ProxyDynamicLease
	ProviderConfig  accountproxy.Config
	ResolveGateways ProviderSessionGatewaysResolver
}

func FetchLeaseProviderSession(ctx context.Context, input ProviderSessionFetchInput) ([]provider.Node, error) {
	providerCfg := input.ProviderConfig
	if input.ResolveGateways != nil {
		gateways, err := input.ResolveGateways(ctx, providerCfg.ProviderID)
		if err != nil {
			return nil, err
		}
		providerCfg.Gateways = gateways
	}
	providerClient, err := NewSessionProvider(input.Factory, providerCfg)
	if err != nil {
		return nil, err
	}
	return FetchProviderSession(ctx, providerClient, input.Lease.GetSession())
}
