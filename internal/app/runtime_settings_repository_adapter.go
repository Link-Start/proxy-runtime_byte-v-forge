package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsRepositoryAdapter struct {
	repository runtimeSettingsRepository
}

func (r runtimeSettingsRepositoryAdapter) View(ctx context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.view(ctx)
}

func (r runtimeSettingsRepositoryAdapter) UpdateDynamicIPProviders(ctx context.Context, providers []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.updateDynamicIPProviders(ctx, providers)
}

func (r runtimeSettingsRepositoryAdapter) Load(ctx context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.load(ctx)
}

func (r runtimeSettingsRepositoryAdapter) Update(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.update(ctx, req)
}

func (r runtimeSettingsRepositoryAdapter) UpdateEgressProfiles(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.updateEgressProfiles(ctx, profiles)
}

func (r runtimeSettingsRepositoryAdapter) UpdateIngressRules(ctx context.Context, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.updateIngressRules(ctx, rules)
}

func (r runtimeSettingsRepositoryAdapter) UpdateInUserRules(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	if r.repository == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return r.repository.updateInUserRules(ctx, profiles, rules)
}
