package store

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func SeedStoreFromConfig(ctx context.Context, accounts ProviderAccountStore, defaultProviderID string, cfg config.Config) error {
	defaultProviderID = strings.TrimSpace(defaultProviderID)
	if defaultProviderID == "" {
		return nil
	}
	existing, err := accounts.ListProviderAccounts(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 || strings.TrimSpace(cfg.Ten24.Username) == "" || strings.TrimSpace(cfg.Ten24.Password) == "" {
		return nil
	}
	_, err = accounts.UpsertProviderAccount(ctx, &proxyruntimev1.UpsertProxyProviderAccountRequest{
		AccountId:     "default-1024proxy",
		ProviderId:    defaultProviderID,
		DisplayName:   "Default 1024Proxy",
		Enabled:       true,
		Username:      cfg.Ten24.Username,
		PasswordValue: cfg.Ten24.Password,
		ClearPassword: false,
	})
	return err
}
