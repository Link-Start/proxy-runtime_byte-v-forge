package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	before, err := repository.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := repository.update(ctx, req)
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after update failed"))
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	settings, err := repository.updateDynamicIPProviders(ctx, req.GetDynamicIpProviders())
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(nil)
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	before, err := repository.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := repository.updateEgressProfiles(ctx, req.GetEgressProfiles())
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after egress profile update failed"))
	return &proxyruntimev1.UpdateProxyEgressProfilesResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	before, err := repository.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := repository.updateIngressRules(ctx, req.GetIngressRules())
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after ingress rule update failed"))
	return &proxyruntimev1.UpdateProxyIngressRulesResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	before, err := repository.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := repository.updateInUserRules(ctx, req.GetEgressProfiles(), req.GetIngressRules())
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after in-user rule update failed"))
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) changedInUserConnectionUsernamesAfterUpdate(ctx context.Context, before *runtimeSettingsFile, errorMessage string) []string {
	repository, err := a.settingsRepository()
	if err != nil {
		a.warn(errorMessage, "error", err)
		return nil
	}
	after, err := repository.load(ctx)
	if err != nil {
		a.warn(errorMessage, "error", err)
		return nil
	}
	return changedInUserConnectionUsernames(before, after)
}
