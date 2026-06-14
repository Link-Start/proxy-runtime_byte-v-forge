package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) releaseLeaseProviderSession(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil || lease.GetSession() == nil || strings.TrimSpace(lease.GetProviderAccountId()) == "" {
		return nil
	}
	if leaseapp.StatelessProviderSession(lease.GetSession()) {
		return nil
	}
	providerCfg, _, err := c.deps.store.ProviderConfig(ctx, lease.GetProviderAccountId())
	if err != nil {
		return err
	}
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return err
	}
	providerCfg.Gateways = endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerCfg.ProviderID)
	providerClient, err := c.newSessionProvider(providerCfg)
	if err != nil {
		return err
	}
	return leaseapp.ReleaseProviderSession(ctx, providerClient, lease.GetSession())
}
