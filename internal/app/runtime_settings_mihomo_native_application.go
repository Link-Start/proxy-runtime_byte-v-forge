package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) GetProxyRuntimeMihomoNativeConfig(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	config, err := mihomoNativeSettings(ctx, a.runtime)
	if err != nil {
		return nil, internalError("load mihomo native config", err)
	}
	return &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse{Config: config}, nil
}

func (a runtimeSettingsApplication) UpdateProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	config, err := updateMihomoNativeSettings(ctx, a.runtime, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse{Config: config}, nil
}
