package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) GetProxyRuntimeSettings(ctx context.Context) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	settings, err := repository.view(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyRuntimeSettingsResponse{Settings: settings}, nil
}
