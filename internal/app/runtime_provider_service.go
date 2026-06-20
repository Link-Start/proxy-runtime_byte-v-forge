package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *RuntimeService) ListProxyProviders(ctx context.Context, _ *proxyruntimev1.ListProxyProvidersRequest) (*proxyruntimev1.ListProxyProvidersResponse, error) {
	return s.providers.ListProxyProviders(ctx)
}

func (s *RuntimeService) ListProxyProviderAccounts(ctx context.Context, _ *proxyruntimev1.ListProxyProviderAccountsRequest) (*proxyruntimev1.ListProxyProviderAccountsResponse, error) {
	return s.providers.ListProxyProviderAccounts(ctx)
}

func (s *RuntimeService) UpsertProxyProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.UpsertProxyProviderAccountResponse, error) {
	return s.providers.UpsertProxyProviderAccount(ctx, req)
}

func (s *RuntimeService) DeleteProxyProviderAccount(ctx context.Context, req *proxyruntimev1.DeleteProxyProviderAccountRequest) (*proxyruntimev1.DeleteProxyProviderAccountResponse, error) {
	return s.providers.DeleteProxyProviderAccount(ctx, req)
}
