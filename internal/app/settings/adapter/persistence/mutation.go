package persistence

import (
	"context"
	"sync"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	settingsdomain "github.com/byte-v-forge/proxy-runtime/internal/app/settings/domain"
)

type runtimeSettingsMutation func(*proxyruntimev1.ProxyRuntimePersistentSettings) (*proxyruntimev1.ProxyRuntimePersistentSettings, error)
type runtimeSettingsChangeMutation func(*proxyruntimev1.ProxyRuntimePersistentSettings) (bool, error)
type runtimeSettingsLoadFunc func(context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error)
type runtimeSettingsSaveFunc func(context.Context, *proxyruntimev1.ProxyRuntimePersistentSettings) error

type runtimeSettingsMutationExecutor struct {
	mu   *sync.Mutex
	load runtimeSettingsLoadFunc
	save runtimeSettingsSaveFunc
}

func (s *Store) mutateRuntimeSettings(ctx context.Context, mutation runtimeSettingsMutation) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutationExecutor().mutate(ctx, mutation)
}

func (s *Store) mutateRuntimeSettingsIfChanged(ctx context.Context, mutation runtimeSettingsChangeMutation) (bool, error) {
	return s.mutationExecutor().mutateIfChanged(ctx, mutation)
}

func (s *Store) mutationExecutor() runtimeSettingsMutationExecutor {
	return runtimeSettingsMutationExecutor{mu: &s.mu, load: s.loadLocked, save: s.saveLocked}
}

func (e runtimeSettingsMutationExecutor) mutate(ctx context.Context, mutation runtimeSettingsMutation) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	settings, err := e.load(ctx)
	if err != nil {
		return nil, err
	}
	next, err := mutation(settings)
	if err != nil {
		return nil, err
	}
	if err := e.save(ctx, next); err != nil {
		return nil, err
	}
	return settingsdomain.RuntimeSettingsView(next), nil
}

func (e runtimeSettingsMutationExecutor) mutateIfChanged(ctx context.Context, mutation runtimeSettingsChangeMutation) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	settings, err := e.load(ctx)
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
	if err := e.save(ctx, settings); err != nil {
		return false, err
	}
	return true, nil
}
