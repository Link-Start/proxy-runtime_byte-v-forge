package lease

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type ProviderSessionGatewaysResolver func(context.Context, string) ([]accountproxy.Gateway, error)

type ProviderSessionReleaseInput struct {
	Store           OrchestrationStore
	Factory         SessionProviderFactory
	Lease           *proxyruntimev1.ProxyDynamicLease
	ResolveGateways ProviderSessionGatewaysResolver
}

func ReleaseLeaseProviderSession(ctx context.Context, input ProviderSessionReleaseInput) error {
	if !NeedsProviderSessionRelease(input.Lease) {
		return nil
	}
	providerCfg, _, err := ProviderConfigForLease(ctx, input.Store, input.Lease)
	if err != nil {
		return err
	}
	if input.ResolveGateways != nil {
		gateways, err := input.ResolveGateways(ctx, providerCfg.ProviderID)
		if err != nil {
			return err
		}
		providerCfg.Gateways = gateways
	}
	providerClient, err := NewSessionProvider(input.Factory, providerCfg)
	if err != nil {
		return err
	}
	return ReleaseProviderSession(ctx, providerClient, input.Lease.GetSession())
}

func NeedsProviderSessionRelease(lease *proxyruntimev1.ProxyDynamicLease) bool {
	if lease == nil || lease.GetSession() == nil || strings.TrimSpace(lease.GetProviderAccountId()) == "" {
		return false
	}
	return !StatelessProviderSession(lease.GetSession())
}

type ProviderSessionReleaseAction func(context.Context, *proxyruntimev1.ProxyDynamicLease) error

func ReleaseLeaseProviderSessionWithLock(ctx context.Context, locks LockManager, lease *proxyruntimev1.ProxyDynamicLease, release ProviderSessionReleaseAction) error {
	if release == nil {
		return nil
	}
	return WithProviderAccountLock(ctx, locks, lease.GetProviderAccountId(), func(ctx context.Context) error {
		return release(ctx, lease)
	})
}
