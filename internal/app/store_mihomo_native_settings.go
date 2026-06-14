package app

import (
	"context"
	"errors"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
	"github.com/jackc/pgx/v5"
)

const mihomoNativeSettingsKey = "mihomo_native"

func (s *PostgresStore) LoadMihomoNativeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	var raw string
	err := s.pool.QueryRow(ctx, `SELECT setting_json::text FROM proxy_runtime_settings WHERE setting_key=$1`, mihomoNativeSettingsKey).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return normalizeMihomoNativeSettings(nil), nil
	}
	if err != nil {
		return nil, err
	}
	return decodeMihomoNativeSettings(raw)
}

func decodeMihomoNativeSettings(raw string) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	settings := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode mihomo native settings: %w", err)
		}
	}
	return normalizeMihomoNativeSettings(settings), nil
}

func (s *PostgresStore) SaveMihomoNativeSettings(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	data, err := protojsoncodec.Marshal(normalizeMihomoNativeSettings(settings))
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO proxy_runtime_settings (setting_key, setting_json) VALUES ($1,$2::jsonb) ON CONFLICT (setting_key) DO UPDATE SET setting_json=EXCLUDED.setting_json, updated_at=now()`, mihomoNativeSettingsKey, string(data))
	return err
}
