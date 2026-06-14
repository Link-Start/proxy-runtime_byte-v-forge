package settings

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a Application) UpdateRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := a.validateProfiles(req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after update failed", func(repository Repository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.Update(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a Application) UpdateEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	if err := a.validateProfiles(req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after egress profile update failed", func(repository Repository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.UpdateEgressProfiles(ctx, req.GetEgressProfiles())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyEgressProfilesResponse{Settings: settings}, nil
}

func (a Application) UpdateIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after ingress rule update failed", func(repository Repository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.UpdateIngressRules(ctx, req.GetIngressRules())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyIngressRulesResponse{Settings: settings}, nil
}

func (a Application) UpdateInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := a.validateProfiles(req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after in-user rule update failed", func(repository Repository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.UpdateInUserRules(ctx, req.GetEgressProfiles(), req.GetIngressRules())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}
