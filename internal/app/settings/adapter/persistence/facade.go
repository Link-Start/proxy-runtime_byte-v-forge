package persistence

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func (s *Store) Load(ctx context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked(ctx)
}

func (s *Store) loadLocked(ctx context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error) {
	if s.store == nil {
		return kernel.NormalizeRuntimeSettingsWithProviders(nil, s.ipFraudProviders, s.ipGeoProviders), nil
	}
	settings, err := s.store.LoadRuntimeSettings(ctx)
	if err != nil {
		return nil, err
	}
	settings = kernel.NormalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders)
	return settings, nil
}

func (s *Store) saveLocked(ctx context.Context, settings *proxyruntimev1.ProxyRuntimePersistentSettings) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveRuntimeSettings(ctx, kernel.NormalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders))
}

func (s *Store) replace(ctx context.Context, settings *proxyruntimev1.ProxyRuntimePersistentSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(ctx, settings)
}
