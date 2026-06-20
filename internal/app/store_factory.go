package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"

	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store/postgres"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store/sqlite"
)

func NewControlStore(ctx context.Context, cfg config.Config, accountProviders *providerregistry.Registry, logger *slog.Logger, clk clock.Clock) (*store.RuntimeStores, error) {
	if strings.TrimSpace(cfg.PostgresDSN) != "" {
		backend, err := postgres.New(ctx, cfg, accountProviders, logger, clk)
		if err != nil {
			return nil, err
		}
		return store.NewRuntimeStores(backend), nil
	}
	backend, err := sqlite.New(ctx, cfg, accountProviders, logger, clk)
	if err != nil {
		return nil, err
	}
	return store.NewRuntimeStores(backend), nil
}
