package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretbox"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db               *sql.DB
	box              secretbox.Box
	accountProviders *providerregistry.Registry
	logger           *slog.Logger
}

func NewSQLiteStore(ctx context.Context, cfg config.Config, accountProviders *providerregistry.Registry, logger *slog.Logger) (*SQLiteStore, error) {
	box, err := secretbox.New(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	dsn, err := sqliteDSN(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &SQLiteStore{db: db, box: box, accountProviders: accountProviders, logger: logger}
	if err := store.applySchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := seedStoreFromConfig(ctx, store, cfg); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func sqliteDSN(dataDir string) (string, error) {
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" {
		return "", fmt.Errorf("PROXY_RUNTIME_DATA_DIR is required")
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dataDir, "proxy-runtime.sqlite")
	uri := &url.URL{Scheme: "file", Path: path}
	query := uri.Query()
	query.Add("_pragma", "busy_timeout(5000)")
	query.Add("_pragma", "journal_mode(WAL)")
	query.Add("_pragma", "foreign_keys(ON)")
	uri.RawQuery = query.Encode()
	return uri.String(), nil
}

func (s *SQLiteStore) Close() {
	if s != nil && s.db != nil {
		_ = s.db.Close()
	}
}
