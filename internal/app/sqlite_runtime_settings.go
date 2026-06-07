package app

import (
	"context"
	"database/sql"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/protojsonx"
)

func (s *SQLiteStore) LoadRuntimeSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT setting_json FROM proxy_runtime_settings WHERE setting_key=?`, runtimeSettingsKey).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return normalizeRuntimeSettings(nil), nil
	}
	if err != nil {
		return nil, err
	}
	settings := &proxyruntimev1.ProxyRuntimePersistentSettings{}
	if raw != "" {
		_ = protojsonx.Unmarshal([]byte(raw), settings)
	}
	return normalizeRuntimeSettings(settings), nil
}

func (s *SQLiteStore) SaveRuntimeSettings(ctx context.Context, settings *runtimeSettingsFile) error {
	data, err := protojsonx.Marshal(normalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	now := sqliteTime(time.Now().UTC())
	_, err = s.db.ExecContext(ctx, `INSERT INTO proxy_runtime_settings (setting_key, setting_json, updated_at) VALUES (?,?,?) ON CONFLICT(setting_key) DO UPDATE SET setting_json=excluded.setting_json, updated_at=excluded.updated_at`, runtimeSettingsKey, string(data), now)
	return err
}
