package app

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

const mihomoNativeSettingsKey = "mihomo_native"

func (s *PostgresStore) LoadMihomoNativeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, mihomoNativeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return normalizeMihomoNativeSettings(nil), nil
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
	return s.saveRuntimeSettingJSON(ctx, mihomoNativeSettingsKey, data)
}
