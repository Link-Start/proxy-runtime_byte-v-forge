package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
)

func (s *runtimeSettingsStore) LoadMihomoNative(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadMihomoNativeLocked(ctx)
}

func (s *runtimeSettingsStore) loadMihomoNativeLocked(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if s.store == nil {
		return mihomonative.NormalizeSettings(nil), nil
	}
	settings, err := s.store.LoadMihomoNativeSettings(ctx)
	if err != nil {
		return nil, err
	}
	return mihomonative.NormalizeSettings(settings), nil
}

func (s *runtimeSettingsStore) SaveMihomoNative(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveMihomoNativeLocked(ctx, settings)
}

func (s *runtimeSettingsStore) saveMihomoNativeLocked(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveMihomoNativeSettings(ctx, mihomonative.NormalizeSettings(settings))
}
