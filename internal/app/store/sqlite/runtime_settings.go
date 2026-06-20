package sqlite

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *Store) LoadRuntimeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, store.RuntimeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return settingscore.NormalizeRuntimeSettings(nil), nil
	}
	return settingscore.DecodeRuntimeSettings(raw)
}

func (s *Store) SaveRuntimeSettings(ctx context.Context, settings *proxyruntimev1.ProxyRuntimePersistentSettings) error {
	data, err := protojsoncodec.Marshal(settingscore.NormalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, store.RuntimeSettingsKey, data)
}
