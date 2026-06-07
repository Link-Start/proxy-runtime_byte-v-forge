package app

import (
	"context"
	"log/slog"
	"time"

	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
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
}

type providerCredential struct {
	Username          string              `json:"username"`
	Password          string              `json:"password,omitempty"`
	PasswordValue     string              `json:"password_value,omitempty"`
	PasswordSecretRef *commonv1.SecretRef `json:"password_secret_ref,omitempty"`
}

type providerAccountRecord struct {
	AccountID         string
	ProviderID        string
	DynamicProviderID string
	DisplayName       string
	Enabled           bool
	CredentialSecret  string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewPostgresStore(ctx context.Context, cfg config.Config, accountProviders *providerregistry.Registry, logger *slog.Logger) (*PostgresStore, error) {
	box, err := secretbox.New(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	store := &PostgresStore{pool: pool, box: box, accountProviders: accountProviders, logger: logger}
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
