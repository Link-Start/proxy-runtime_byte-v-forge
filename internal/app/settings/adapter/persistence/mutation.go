package persistence

import (
	"context"
	"sync"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

type runtimeSettingsMutation func(*proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
type runtimeSettingsChangeMutation func(*proxygatewayv1.ProxyGatewayPersistentSettings) (bool, error)
type runtimeSettingsLoadFunc func(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
type runtimeSettingsSaveFunc func(context.Context, *proxygatewayv1.ProxyGatewayPersistentSettings) error

type runtimeSettingsMutationExecutor struct {
	mu   *sync.Mutex
	load runtimeSettingsLoadFunc
	save runtimeSettingsSaveFunc
}

func (s *Store) mutateRuntimeSettings(ctx context.Context, mutation runtimeSettingsMutation) (*proxygatewayv1.ProxyGatewaySettings, error) {
	return s.mutationExecutor().mutate(ctx, mutation)
}

func (s *Store) mutateRuntimeSettingsIfChanged(ctx context.Context, mutation runtimeSettingsChangeMutation) (bool, error) {
	return s.mutationExecutor().mutateIfChanged(ctx, mutation)
}

func (s *Store) mutationExecutor() runtimeSettingsMutationExecutor {
	return runtimeSettingsMutationExecutor{mu: &s.mu, load: s.loadLocked, save: s.saveLocked}
}

func (e runtimeSettingsMutationExecutor) mutate(ctx context.Context, mutation runtimeSettingsMutation) (*proxygatewayv1.ProxyGatewaySettings, error) {
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
