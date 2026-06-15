package app

import (
	"context"
	"fmt"
)

const sqliteSchemaSQL = `
CREATE TABLE IF NOT EXISTS proxy_runtime_provider_accounts (
  account_id text PRIMARY KEY,
  provider_id text NOT NULL,
  dynamic_provider_id text NOT NULL DEFAULT '',
  display_name text NOT NULL,
  enabled integer NOT NULL DEFAULT 1,
  credential_secret text NOT NULL DEFAULT '',
  created_at text NOT NULL,
  updated_at text NOT NULL
);

CREATE TABLE IF NOT EXISTS proxy_runtime_dynamic_leases (
  lease_id text PRIMARY KEY,
  account_id text NOT NULL,
  purpose text NOT NULL DEFAULT '',
  provider_account_id text NOT NULL DEFAULT '',
  status text NOT NULL,
  lease_json text NOT NULL,
  acquired_at text NOT NULL DEFAULT '',
  expires_at text NOT NULL DEFAULT '',
  created_at text NOT NULL,
  updated_at text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_account_status
  ON proxy_runtime_dynamic_leases(account_id, status);
CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_account_status_expires
  ON proxy_runtime_dynamic_leases(account_id, status, expires_at);
CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_provider_status
  ON proxy_runtime_dynamic_leases(provider_account_id, status);
CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_provider_status_expires
  ON proxy_runtime_dynamic_leases(provider_account_id, status, expires_at);
CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_status_expires
  ON proxy_runtime_dynamic_leases(status, expires_at);
CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_expires_at
  ON proxy_runtime_dynamic_leases(expires_at);
CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_acquired_updated
  ON proxy_runtime_dynamic_leases(acquired_at DESC, updated_at DESC, lease_id);


CREATE TABLE IF NOT EXISTS proxy_runtime_secrets (
  secret_id text PRIMARY KEY,
  provider text NOT NULL,
  purpose text NOT NULL,
  secret_payload text NOT NULL,
  expires_at text NOT NULL DEFAULT '',
  created_at text NOT NULL,
  updated_at text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_proxy_runtime_secrets_expires_at
  ON proxy_runtime_secrets(expires_at);

CREATE TABLE IF NOT EXISTS proxy_runtime_settings (
  setting_key text PRIMARY KEY,
  setting_json text NOT NULL DEFAULT '{}',
  updated_at text NOT NULL
);
`

func (s *SQLiteStore) applySchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, sqliteSchemaSQL); err != nil {
		return fmt.Errorf("apply proxy-runtime sqlite schema: %w", err)
	}
	return nil
}
