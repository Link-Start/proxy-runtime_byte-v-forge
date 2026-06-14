package settings

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrMihomoNativeUpdaterRequired = errors.New("mihomo native settings updater is required")

type MihomoNativeLoader func(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
type MihomoNativeUpdater func(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
type MihomoNativeDefault func() *proxyruntimev1.ProxyRuntimeMihomoNativeConfig

func (a Application) GetMihomoNative(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	config, err := a.loadMihomoNative(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse{Config: config}, nil
}

func (a Application) UpdateMihomoNative(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	config, err := a.updateMihomoNative(ctx, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse{Config: config}, nil
}

func (a Application) loadMihomoNative(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if a.loadMihomoNativeSettings == nil {
		return a.defaultMihomoNative(), nil
	}
	return a.loadMihomoNativeSettings(ctx)
}

func (a Application) updateMihomoNative(ctx context.Context, config *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if a.updateMihomoNativeSettings == nil {
		return nil, ErrMihomoNativeUpdaterRequired
	}
	return a.updateMihomoNativeSettings(ctx, config)
}

func (a Application) defaultMihomoNative() *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	if a.defaultMihomoNativeSettings == nil {
		return &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	return a.defaultMihomoNativeSettings()
}
