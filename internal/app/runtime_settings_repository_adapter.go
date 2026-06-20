package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

type runtimeSettingsRepositoryAdapter struct {
	repository runtimeSettingsRepository
}

func (r runtimeSettingsRepositoryAdapter) repositoryOrError() (runtimeSettingsRepository, error) {
	if r.repository == nil {
		return nil, appcore.InternalError("runtime settings repository is not configured", nil)
	}
	return r.repository, nil
}

func (r runtimeSettingsRepositoryAdapter) View(ctx context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.view(ctx)
}

func (r runtimeSettingsRepositoryAdapter) UpdateDynamicIPProviders(ctx context.Context, providers []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.updateDynamicIPProviders(ctx, providers)
}

func (r runtimeSettingsRepositoryAdapter) Load(ctx context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.load(ctx)
}

func (r runtimeSettingsRepositoryAdapter) Update(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.update(ctx, req)
}

func (r runtimeSettingsRepositoryAdapter) UpdateEgressProfiles(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.updateEgressProfiles(ctx, profiles)
}

func (r runtimeSettingsRepositoryAdapter) UpdateIngressRules(ctx context.Context, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.updateIngressRules(ctx, rules)
}

func (r runtimeSettingsRepositoryAdapter) UpdateInUserRules(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := r.repositoryOrError()
	if err != nil {
		return nil, err
	}
	return repository.updateInUserRules(ctx, profiles, rules)
}
