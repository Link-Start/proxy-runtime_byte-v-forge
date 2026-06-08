package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (s *PostgresStore) seedFromConfig(ctx context.Context, cfg config.Config) error {
	return seedStoreFromConfig(ctx, s, cfg)
}

func seedStoreFromConfig(ctx context.Context, store controlStore, cfg config.Config) error {
	accounts, err := store.ListProviderAccounts(ctx)
	if err != nil {
		return err
	}
	if len(accounts) > 0 || strings.TrimSpace(cfg.Ten24.Username) == "" || strings.TrimSpace(cfg.Ten24.Password) == "" {
		return nil
	}
	_, err = store.UpsertProviderAccount(ctx, &proxyruntimev1.UpsertProxyProviderAccountRequest{
		AccountId:     "default-1024proxy",
		ProviderId:    accountproxy.ProviderTen24,
		DisplayName:   "Default 1024Proxy",
		Enabled:       true,
		Username:      cfg.Ten24.Username,
		PasswordValue: cfg.Ten24.Password,
		ClearPassword: false,
	})
	return err
}
