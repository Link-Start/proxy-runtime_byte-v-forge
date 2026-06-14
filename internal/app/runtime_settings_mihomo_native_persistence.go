package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *runtimeSettingsStore) loadMihomoNative(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadMihomoNativeLocked(ctx)
}

func (s *runtimeSettingsStore) loadMihomoNativeLocked(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if s.store == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	settings, err := s.store.LoadMihomoNativeSettings(ctx)
	if err != nil {
		return nil, err
	}
	return normalizeMihomoNativeSettings(settings), nil
}

func (s *runtimeSettingsStore) saveMihomoNative(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveMihomoNativeLocked(ctx, settings)
}

func (s *runtimeSettingsStore) saveMihomoNativeLocked(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveMihomoNativeSettings(ctx, normalizeMihomoNativeSettings(settings))
}
