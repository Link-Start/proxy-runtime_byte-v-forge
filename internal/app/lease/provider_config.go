package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func ProviderConfigForLease(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease) (accountproxy.Config, string, error) {
	if store == nil {
		return accountproxy.Config{}, "", errors.New("lease store is required")
	}
	return store.ProviderConfig(ctx, lease.GetProviderAccountId())
}

func ProviderConfigForGateway(ctx context.Context, store OrchestrationStore, providerAccountID string, gateway accountproxy.Gateway) (accountproxy.Config, string, error) {
	if store == nil {
		return accountproxy.Config{}, "", errors.New("lease store is required")
	}
	providerCfg, accountID, err := store.ProviderConfig(ctx, providerAccountID)
	if err != nil {
		return accountproxy.Config{}, "", err
	}
	providerCfg.Gateways = []accountproxy.Gateway{gateway}
	return providerCfg, accountID, nil
}
