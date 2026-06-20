package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func (s *runtimeSettingsStore) load(ctx context.Context) (*runtimeSettingsFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked(ctx)
}

func (s *runtimeSettingsStore) loadLocked(ctx context.Context) (*runtimeSettingsFile, error) {
	if s.store == nil {
		return settingscore.NormalizeRuntimeSettingsWithProviders(nil, s.ipFraudProviders, s.ipGeoProviders), nil
	}
	settings, err := s.store.LoadRuntimeSettings(ctx)
	if err != nil {
		return nil, err
	}
	settings = settingscore.NormalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders)
	return settings, nil
}

func (s *runtimeSettingsStore) saveLocked(ctx context.Context, settings *runtimeSettingsFile) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveRuntimeSettings(ctx, settingscore.NormalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders))
}

func (s *runtimeSettingsStore) replace(ctx context.Context, settings *runtimeSettingsFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(ctx, settings)
}
