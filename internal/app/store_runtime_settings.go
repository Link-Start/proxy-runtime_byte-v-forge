package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

const runtimeSettingsKey = "runtime"

func (s *PostgresStore) LoadRuntimeSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, runtimeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return normalizeRuntimeSettings(nil), nil
	}
	return decodeRuntimeSettings(raw)
}

func (s *PostgresStore) SaveRuntimeSettings(ctx context.Context, settings *runtimeSettingsFile) error {
	data, err := protojsoncodec.Marshal(normalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, runtimeSettingsKey, data)
}
