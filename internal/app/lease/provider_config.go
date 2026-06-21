package lease

import (
	"context"
	"errors"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"
)

func ProviderConfigForLease(ctx context.Context, store OrchestrationStore, lease *proxygatewayv1.ProxyDynamicLease) (accountproxy.Config, string, error) {
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
