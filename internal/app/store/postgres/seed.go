package postgres

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/config"

	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *Store) seedFromConfig(ctx context.Context, cfg config.Config) error {
	return store.SeedStoreFromConfig(ctx, s, s.accountProviders.DefaultProviderID(), cfg)
}
