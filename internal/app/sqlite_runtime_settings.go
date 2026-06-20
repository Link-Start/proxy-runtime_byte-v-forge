package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func (s *SQLiteStore) LoadRuntimeSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, runtimeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return settingscore.NormalizeRuntimeSettings(nil), nil
	}
	return settingscore.DecodeRuntimeSettings(raw)
}

func (s *SQLiteStore) SaveRuntimeSettings(ctx context.Context, settings *runtimeSettingsFile) error {
	data, err := protojsoncodec.Marshal(settingscore.NormalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, runtimeSettingsKey, data)
}
