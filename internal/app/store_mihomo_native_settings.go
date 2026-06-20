package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *PostgresStore) LoadMihomoNativeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, store.MihomoNativeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return mihomonative.NormalizeSettings(nil), nil
	}
	return settingscore.DecodeMihomoNativeSettings(raw)
}

func (s *PostgresStore) SaveMihomoNativeSettings(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	data, err := protojsoncodec.Marshal(mihomonative.NormalizeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, store.MihomoNativeSettingsKey, data)
}
