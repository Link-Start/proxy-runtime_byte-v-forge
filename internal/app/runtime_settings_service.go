package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *RuntimeService) ListProxyIPFraudProviders(ctx context.Context, _ *proxyruntimev1.ListProxyIPFraudProvidersRequest) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return s.settings.ListProxyIPFraudProviders(ctx)
}

func (s *RuntimeService) ListProxyIPGeoProviders(ctx context.Context, _ *proxyruntimev1.ListProxyIPGeoProvidersRequest) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return s.settings.ListProxyIPGeoProviders(ctx)
}

func (s *RuntimeService) GetProxyRuntimeSettings(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeSettingsRequest) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	return s.settings.GetProxyRuntimeSettings(ctx)
}

func (s *RuntimeService) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyRuntimeSettings(ctx, req)
}

func (s *RuntimeService) UpdateProxyEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	return s.settings.UpdateProxyEgressProfiles(ctx, req)
}

func (s *RuntimeService) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	return s.settings.UpdateProxyIngressRules(ctx, req)
}

func (s *RuntimeService) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyDynamicIPProviders(ctx, req)
}

func (s *RuntimeService) UpdateProxyInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyInUserRules(ctx, req)
}

func (s *RuntimeService) GetProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.GetProxyRuntimeMihomoNativeConfigResponse, error) {
	return s.settings.GetProxyRuntimeMihomoNativeConfig(ctx, req)
}

func (s *RuntimeService) UpdateProxyRuntimeMihomoNativeConfig(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest) (*proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigResponse, error) {
	return s.settings.UpdateProxyRuntimeMihomoNativeConfig(ctx, req)
}
