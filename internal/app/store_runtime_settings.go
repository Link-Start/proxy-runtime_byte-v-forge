package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
	"github.com/jackc/pgx/v5"
)

const runtimeSettingsKey = "runtime"

func (s *PostgresStore) LoadRuntimeSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	var raw string
	err := s.pool.QueryRow(ctx, `SELECT setting_json::text FROM proxy_runtime_settings WHERE setting_key=$1`, runtimeSettingsKey).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return normalizeRuntimeSettings(nil), nil
	}
	if err != nil {
		return nil, err
	}
	return decodeRuntimeSettings(raw)
}

func decodeRuntimeSettings(raw string) (*runtimeSettingsFile, error) {
	settings := &runtimeSettingsFile{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode runtime settings: %w", err)
		}
	}
	return normalizeRuntimeSettings(settings), nil
}

func (s *PostgresStore) SaveRuntimeSettings(ctx context.Context, settings *runtimeSettingsFile) error {
	data, err := protojsoncodec.Marshal(normalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO proxy_runtime_settings (setting_key, setting_json) VALUES ($1,$2::jsonb) ON CONFLICT (setting_key) DO UPDATE SET setting_json=EXCLUDED.setting_json, updated_at=now()`, runtimeSettingsKey, string(data))
	return err
}
