package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Store) loadRuntimeSettingJSON(ctx context.Context, key string) (string, bool, error) {
	var raw string
	err := s.pool.QueryRow(ctx, `SELECT setting_json::text FROM proxy_gateway_settings WHERE setting_key=$1`, key).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return raw, true, nil
}

func (s *Store) saveRuntimeSettingJSON(ctx context.Context, key string, data []byte) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO proxy_gateway_settings (setting_key, setting_json) VALUES ($1,$2::jsonb) ON CONFLICT (setting_key) DO UPDATE SET setting_json=EXCLUDED.setting_json, updated_at=now()`, key, string(data))
	return err
}
