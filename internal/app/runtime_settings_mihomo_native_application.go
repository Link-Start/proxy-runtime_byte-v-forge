package app

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings"
)

func (a runtimeSettingsApplication) GetProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	response, err := a.settingsUsecase().GetMihomoNative(ctx, req)
	if err != nil {
		return nil, internalError("load mihomo native config", err)
	}
	return response, nil
}

func (a runtimeSettingsApplication) UpdateProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	response, err := a.settingsUsecase().UpdateMihomoNative(ctx, req)
	if err != nil {
		if errors.Is(err, settingsapp.ErrMihomoNativeUpdaterRequired) {
			return nil, internalError(err.Error(), err)
		}
		return nil, err
	}
	return response, nil
}
