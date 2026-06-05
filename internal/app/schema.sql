CREATE TABLE IF NOT EXISTS proxy_runtime_provider_accounts (
  account_id text PRIMARY KEY,
  provider_id text NOT NULL,
  display_name text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  credential_secret text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS proxy_runtime_sources (
  source_id text PRIMARY KEY,
  source_kind text NOT NULL,
  display_name text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  source_secret text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_proxy_runtime_sources_kind_enabled
  ON proxy_runtime_sources(source_kind, enabled);

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

CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_provider_status
  ON proxy_runtime_dynamic_leases(provider_account_id, status);

CREATE INDEX IF NOT EXISTS idx_proxy_runtime_dynamic_leases_expires_at
  ON proxy_runtime_dynamic_leases(expires_at);

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

UPDATE proxy_runtime_settings
SET setting_json = jsonb_set(
  setting_json,
  '{dynamic_ip_providers}',
  COALESCE((
    SELECT jsonb_agg(
      jsonb_set(
        provider,
        '{gateways}',
        COALESCE((
          SELECT jsonb_agg(
            CASE
              WHEN jsonb_typeof(gateway) = 'object' AND gateway ? 'endpoint_url'
                THEN jsonb_build_object('endpoint_url', gateway->'endpoint_url')
              WHEN jsonb_typeof(gateway) = 'object' AND gateway ? 'addr'
                THEN jsonb_build_object('endpoint_url', gateway->'addr')
              ELSE jsonb_build_object('endpoint_url', '')
            END
          )
          FROM jsonb_array_elements(COALESCE(provider->'gateways', '[]'::jsonb)) AS gateway
        ), '[]'::jsonb),
        true
      )
    )
    FROM jsonb_array_elements(setting_json->'dynamic_ip_providers') AS provider
  ), '[]'::jsonb),
  true
)
WHERE setting_key = 'runtime'
  AND jsonb_typeof(setting_json->'dynamic_ip_providers') = 'array';
