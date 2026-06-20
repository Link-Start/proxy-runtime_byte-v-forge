package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *PostgresStore) LoadRuntimeSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, store.RuntimeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return settingscore.NormalizeRuntimeSettings(nil), nil
	}
	return settingscore.DecodeRuntimeSettings(raw)
}

func (s *PostgresStore) SaveRuntimeSettings(ctx context.Context, settings *runtimeSettingsFile) error {
	data, err := protojsoncodec.Marshal(settingscore.NormalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, store.RuntimeSettingsKey, data)
}
