package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *runtimeSettingsStore) load(ctx context.Context) (*runtimeSettingsFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked(ctx)
}

func (s *runtimeSettingsStore) loadLocked(ctx context.Context) (*runtimeSettingsFile, error) {
	if s.store == nil {
		return normalizeRuntimeSettingsWithProviders(nil, s.ipFraudProviders, s.ipGeoProviders), nil
	}
	settings, err := s.store.LoadRuntimeSettings(ctx)
	if err != nil {
		return nil, err
	}
	settings = normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders)
	return settings, nil
}

func (s *runtimeSettingsStore) saveLocked(ctx context.Context, settings *runtimeSettingsFile) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveRuntimeSettings(ctx, normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders))
}

func (s *runtimeSettingsStore) replace(ctx context.Context, settings *runtimeSettingsFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(ctx, settings)
}

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
