package application

import (
	"context"
	"errors"
	"fmt"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

var (
	ErrMihomoNativeUpdateUnavailable = errors.New("mihomo native settings update unavailable")
	errMihomoNativeUpdaterRequired   = errors.New("mihomo native settings updater is required")
)

type MihomoNativeLoader func(context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error)
type MihomoNativeUpdater func(context.Context, *proxygatewayv1.ProxyGatewayMihomoNativeConfig) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error)
type MihomoNativeDefault func() *proxygatewayv1.ProxyGatewayMihomoNativeConfig

func (a Application) GetMihomoNative(ctx context.Context, _ *proxygatewayv1.GetProxyGatewayMihomoNativeConfigRequest) (*proxygatewayv1.GetProxyGatewayMihomoNativeConfigResponse, error) {
	config, err := a.loadMihomoNative(ctx)
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.GetProxyGatewayMihomoNativeConfigResponse{Config: config}, nil
}

func (a Application) UpdateMihomoNative(ctx context.Context, req *proxygatewayv1.UpdateProxyGatewayMihomoNativeConfigRequest) (*proxygatewayv1.UpdateProxyGatewayMihomoNativeConfigResponse, error) {
	config, err := a.updateMihomoNative(ctx, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.UpdateProxyGatewayMihomoNativeConfigResponse{Config: config}, nil
}

func (a Application) loadMihomoNative(ctx context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	if a.loadMihomoNativeSettings == nil {
		return a.defaultMihomoNative(), nil
	}
	return a.loadMihomoNativeSettings(ctx)
}

func (a Application) updateMihomoNative(ctx context.Context, config *proxygatewayv1.ProxyGatewayMihomoNativeConfig) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	if a.updateMihomoNativeSettings == nil {
		return nil, fmt.Errorf("%w: %w", ErrMihomoNativeUpdateUnavailable, errMihomoNativeUpdaterRequired)
	}
	return a.updateMihomoNativeSettings(ctx, config)
}

func (a Application) defaultMihomoNative() *proxygatewayv1.ProxyGatewayMihomoNativeConfig {
	if a.defaultMihomoNativeSettings == nil {
		return &proxygatewayv1.ProxyGatewayMihomoNativeConfig{}
	}
	return a.defaultMihomoNativeSettings()
}
