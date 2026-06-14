package app

import (
	"context"
	"fmt"

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
	return s.saveRuntimeSettingJSON(ctx, runtimeSettingsKey, data)
}
