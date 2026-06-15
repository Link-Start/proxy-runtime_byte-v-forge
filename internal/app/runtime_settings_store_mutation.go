package app

import (
	"context"
	"sync"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsMutation func(*runtimeSettingsFile) (*runtimeSettingsFile, error)
type runtimeSettingsChangeMutation func(*runtimeSettingsFile) (bool, error)
type runtimeSettingsLoadFunc func(context.Context) (*runtimeSettingsFile, error)
type runtimeSettingsSaveFunc func(context.Context, *runtimeSettingsFile) error

type runtimeSettingsMutationExecutor struct {
	mu   *sync.Mutex
	load runtimeSettingsLoadFunc
	save runtimeSettingsSaveFunc
}

func (s *runtimeSettingsStore) mutateRuntimeSettings(ctx context.Context, mutation runtimeSettingsMutation) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutationExecutor().mutate(ctx, mutation)
}

func (s *runtimeSettingsStore) mutateRuntimeSettingsIfChanged(ctx context.Context, mutation runtimeSettingsChangeMutation) (bool, error) {
	return s.mutationExecutor().mutateIfChanged(ctx, mutation)
}

func (s *runtimeSettingsStore) mutationExecutor() runtimeSettingsMutationExecutor {
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
	return runtimeSettingsView(next), nil
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
