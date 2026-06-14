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
