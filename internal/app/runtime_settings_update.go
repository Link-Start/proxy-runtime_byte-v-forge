package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateSettingsWithConnectionCleanup(ctx, "load runtime settings after update failed", func(repository runtimeSettingsRepository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.update(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	settings, err := a.updateSettingsAndScheduleApply(func(repository runtimeSettingsRepository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.updateDynamicIPProviders(ctx, req.GetDynamicIpProviders())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateSettingsWithConnectionCleanup(ctx, "load runtime settings after egress profile update failed", func(repository runtimeSettingsRepository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.updateEgressProfiles(ctx, req.GetEgressProfiles())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyEgressProfilesResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	settings, err := a.updateSettingsWithConnectionCleanup(ctx, "load runtime settings after ingress rule update failed", func(repository runtimeSettingsRepository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.updateIngressRules(ctx, req.GetIngressRules())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyIngressRulesResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateSettingsWithConnectionCleanup(ctx, "load runtime settings after in-user rule update failed", func(repository runtimeSettingsRepository) (*proxyruntimev1.ProxyRuntimeSettings, error) {
		return repository.updateInUserRules(ctx, req.GetEgressProfiles(), req.GetIngressRules())
	})
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}
