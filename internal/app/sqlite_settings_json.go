package app

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s *SQLiteStore) loadRuntimeSettingJSON(ctx context.Context, key string) (string, bool, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT setting_json FROM proxy_runtime_settings WHERE setting_key=?`, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return raw, true, nil
}

func (s *SQLiteStore) saveRuntimeSettingJSON(ctx context.Context, key string, data []byte) error {
	now := sqliteTime(time.Now().UTC())
	_, err := s.db.ExecContext(ctx, `INSERT INTO proxy_runtime_settings (setting_key, setting_json, updated_at) VALUES (?,?,?) ON CONFLICT(setting_key) DO UPDATE SET setting_json=excluded.setting_json, updated_at=excluded.updated_at`, key, string(data), now)
	return err
}
