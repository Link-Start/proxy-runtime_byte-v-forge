CREATE TABLE IF NOT EXISTS proxy_runtime_provider_accounts (
  account_id text PRIMARY KEY,
  provider_id text NOT NULL,
  dynamic_provider_id text NOT NULL DEFAULT '',
  display_name text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  credential_secret text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE proxy_runtime_provider_accounts
  ADD COLUMN IF NOT EXISTS dynamic_provider_id text NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS proxy_runtime_dynamic_leases (
  lease_id text PRIMARY KEY,
  account_id text NOT NULL,
  purpose text NOT NULL DEFAULT '',
  provider_account_id text NOT NULL DEFAULT '',
  status text NOT NULL,
  lease_json jsonb NOT NULL,
  acquired_at timestamptz NULL,
  expires_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
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
  ON proxy_runtime_dynamic_leases(acquired_at DESC NULLS LAST, updated_at DESC, lease_id);


CREATE TABLE IF NOT EXISTS proxy_runtime_secrets (
  secret_id text PRIMARY KEY,
  provider text NOT NULL,
  purpose text NOT NULL,
  secret_payload text NOT NULL,
  expires_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_proxy_runtime_secrets_expires_at
  ON proxy_runtime_secrets(expires_at);

CREATE TABLE IF NOT EXISTS proxy_runtime_settings (
  setting_key text PRIMARY KEY,
  setting_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_at timestamptz NOT NULL DEFAULT now()
);
