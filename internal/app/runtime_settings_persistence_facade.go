package app

import "context"

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
