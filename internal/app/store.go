package app

import (
	"context"
	"log/slog"

	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretbox"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool             *pgxpool.Pool
	box              secretbox.Box
	accountProviders *providerregistry.Registry
	logger           *slog.Logger
	clock            clock.Clock
}

func NewPostgresStore(ctx context.Context, cfg config.Config, accountProviders *providerregistry.Registry, logger *slog.Logger, clk clock.Clock) (*PostgresStore, error) {
	box, err := secretbox.New(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	store := &PostgresStore{pool: pool, box: box, accountProviders: accountProviders, logger: logger, clock: clk}
	if err := store.applySchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := store.seedFromConfig(ctx, cfg); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}
