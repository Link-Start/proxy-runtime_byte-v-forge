package app

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings"
)

func (a runtimeSettingsApplication) GetProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	response, err := a.usecase.GetMihomoNative(ctx, req)
	if err != nil {
		return nil, internalError("load mihomo native config", err)
	}
	return response, nil
}

func (a runtimeSettingsApplication) UpdateProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	response, err := a.usecase.UpdateMihomoNative(ctx, req)
	if err != nil {
		if errors.Is(err, settingsapp.ErrMihomoNativeUpdateUnavailable) {
			return nil, internalError("mihomo native settings update unavailable", err)
		}
		return nil, err
	}
	return response, nil
}
