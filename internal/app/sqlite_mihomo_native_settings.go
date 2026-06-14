package app

import (
	"context"
	"database/sql"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

func (s *SQLiteStore) LoadMihomoNativeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT setting_json FROM proxy_runtime_settings WHERE setting_key=?`, mihomoNativeSettingsKey).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return normalizeMihomoNativeSettings(nil), nil
	}
	if err != nil {
		return nil, err
	}
	return decodeMihomoNativeSettings(raw)
}

func (s *SQLiteStore) SaveMihomoNativeSettings(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	data, err := protojsoncodec.Marshal(normalizeMihomoNativeSettings(settings))
	if err != nil {
		return err
	}
	now := sqliteTime(time.Now().UTC())
	_, err = s.db.ExecContext(ctx, `INSERT INTO proxy_runtime_settings (setting_key, setting_json, updated_at) VALUES (?,?,?) ON CONFLICT(setting_key) DO UPDATE SET setting_json=excluded.setting_json, updated_at=excluded.updated_at`, mihomoNativeSettingsKey, string(data), now)
	return err
}
