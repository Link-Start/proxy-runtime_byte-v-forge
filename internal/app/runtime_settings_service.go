package app

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings/application"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (s *RuntimeService) ListProxyIPFraudProviders(ctx context.Context, _ *proxyruntimev1.ListProxyIPFraudProvidersRequest) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return s.settings.ListIPFraudProviders(ctx)
}

func (s *RuntimeService) ListProxyIPGeoProviders(ctx context.Context, _ *proxyruntimev1.ListProxyIPGeoProvidersRequest) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return s.settings.ListIPGeoProviders(ctx)
}

func (s *RuntimeService) GetProxyRuntimeSettings(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeSettingsRequest) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	return s.settings.Get(ctx)
}

func (s *RuntimeService) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateRuntimeSettings(ctx, req)
}

func (s *RuntimeService) UpdateProxyEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	return s.settings.UpdateEgressProfiles(ctx, req)
}

func (s *RuntimeService) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	return s.settings.UpdateIngressRules(ctx, req)
}

func (s *RuntimeService) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateDynamicIPProviders(ctx, req.GetDynamicIpProviders())
}

func (s *RuntimeService) UpdateProxyInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateInUserRules(ctx, req)
}

func (s *RuntimeService) GetProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	response, err := s.settings.GetMihomoNative(ctx, req)
	if err != nil {
		return nil, appcore.InternalError("load mihomo native config", err)
	}
	return response, nil
}

func (s *RuntimeService) UpdateProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	response, err := s.settings.UpdateMihomoNative(ctx, req)
	if err != nil {
		if errors.Is(err, settingsapp.ErrMihomoNativeUpdateUnavailable) {
			return nil, appcore.InternalError("mihomo native settings update unavailable", err)
		}
		return nil, err
	}
	return response, nil
}
