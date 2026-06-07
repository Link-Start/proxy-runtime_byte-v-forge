package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

func NewControlStore(ctx context.Context, cfg config.Config, accountProviders *providerregistry.Registry, logger *slog.Logger) (controlStore, error) {
	if strings.TrimSpace(cfg.PostgresDSN) != "" {
		return NewPostgresStore(ctx, cfg, accountProviders, logger)
	}
	return NewSQLiteStore(ctx, cfg, accountProviders, logger)
}
