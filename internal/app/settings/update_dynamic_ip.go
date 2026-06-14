package settings

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a Application) UpdateDynamicIPProviders(ctx context.Context, providers []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	settings, err := a.updateAndSchedule(ctx, func(repository Repository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.UpdateDynamicIPProviders(ctx, providers)
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}
