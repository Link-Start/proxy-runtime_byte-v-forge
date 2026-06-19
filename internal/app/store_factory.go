package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

func NewControlStore(ctx context.Context, cfg config.Config, accountProviders *providerregistry.Registry, logger *slog.Logger, clk clock.Clock) (*RuntimeStores, error) {
	if strings.TrimSpace(cfg.PostgresDSN) != "" {
		store, err := NewPostgresStore(ctx, cfg, accountProviders, logger, clk)
		if err != nil {
			return nil, err
		}
		return newRuntimeStores(store), nil
	}
	store, err := NewSQLiteStore(ctx, cfg, accountProviders, logger, clk)
	if err != nil {
		return nil, err
	}
	return newRuntimeStores(store), nil
}
