package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsMutation func(*runtimeSettingsFile) (*runtimeSettingsFile, error)
type runtimeSettingsChangeMutation func(*runtimeSettingsFile) (bool, error)

func (s *runtimeSettingsStore) mutateRuntimeSettings(ctx context.Context, mutation runtimeSettingsMutation) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	next, err := mutation(settings)
	if err != nil {
		return nil, err
	}
	if err := s.saveLocked(ctx, next); err != nil {
		return nil, err
	}
	return runtimeSettingsView(next), nil
}

func (s *runtimeSettingsStore) mutateRuntimeSettingsIfChanged(ctx context.Context, mutation runtimeSettingsChangeMutation) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return false, err
	}
	changed, err := mutation(settings)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	if err := s.saveLocked(ctx, settings); err != nil {
		return false, err
	}
	return true, nil
}
