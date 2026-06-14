package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

func (s *SQLiteStore) LoadMihomoNativeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, mihomoNativeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return normalizeMihomoNativeSettings(nil), nil
	}
	return decodeMihomoNativeSettings(raw)
}

func (s *SQLiteStore) SaveMihomoNativeSettings(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	data, err := protojsoncodec.Marshal(normalizeMihomoNativeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, mihomoNativeSettingsKey, data)
}
