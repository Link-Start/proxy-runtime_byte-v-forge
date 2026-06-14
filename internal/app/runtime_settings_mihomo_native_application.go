package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) GetProxyRuntimeMihomoNativeConfig(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	config, err := a.loadRuntimeMihomoNativeSettings(ctx)
	if err != nil {
		return nil, internalError("load mihomo native config", err)
	}
	return &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse{Config: config}, nil
}

func (a runtimeSettingsApplication) UpdateProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	config, err := a.updateRuntimeMihomoNativeSettings(ctx, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse{Config: config}, nil
}

func (a runtimeSettingsApplication) loadRuntimeMihomoNativeSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if a.loadMihomoNativeSettings == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	return a.loadMihomoNativeSettings(ctx)
}

func (a runtimeSettingsApplication) updateRuntimeMihomoNativeSettings(ctx context.Context, config *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if a.updateMihomoNativeSettings == nil {
		return nil, internalError("mihomo native settings updater is required", nil)
	}
	return a.updateMihomoNativeSettings(ctx, config)
}
