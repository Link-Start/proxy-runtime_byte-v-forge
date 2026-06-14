package app

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) loadRuntimeSettingJSON(ctx context.Context, key string) (string, bool, error) {
	var raw string
	err := s.pool.QueryRow(ctx, `SELECT setting_json::text FROM proxy_runtime_settings WHERE setting_key=$1`, key).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return raw, true, nil
}

func (s *PostgresStore) saveRuntimeSettingJSON(ctx context.Context, key string, data []byte) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO proxy_runtime_settings (setting_key, setting_json) VALUES ($1,$2::jsonb) ON CONFLICT (setting_key) DO UPDATE SET setting_json=EXCLUDED.setting_json, updated_at=now()`, key, string(data))
	return err
}
