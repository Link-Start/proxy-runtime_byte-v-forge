package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type restoreLeaseInputs struct {
	providerConfig  accountproxy.Config
	providerAccount *proxyruntimev1.ProxyProviderAccount
	settings        *runtimeSettingsFile
}

func (c leaseCoordinator) loadRestoreLeaseInputs(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) (restoreLeaseInputs, error) {
	providerCfg, _, err := c.deps.store.ProviderConfig(ctx, lease.GetProviderAccountId())
	if err != nil {
		return restoreLeaseInputs{}, err
	}
	providerAccount, err := c.deps.store.ProviderAccount(ctx, lease.GetProviderAccountId())
	if err != nil {
		return restoreLeaseInputs{}, err
	}
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return restoreLeaseInputs{}, err
	}
	return restoreLeaseInputs{providerConfig: providerCfg, providerAccount: providerAccount, settings: settings}, nil
}
