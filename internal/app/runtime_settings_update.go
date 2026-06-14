package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.proxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	before, err := a.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := a.settings.update(ctx, req)
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after update failed"))
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	settings, err := a.settings.updateDynamicIPProviders(ctx, req.GetDynamicIpProviders())
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
	before, err := a.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := a.settings.updateEgressProfiles(ctx, req.GetEgressProfiles())
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after egress profile update failed"))
	return &proxyruntimev1.UpdateProxyEgressProfilesResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	before, err := a.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := a.settings.updateIngressRules(ctx, req.GetIngressRules())
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
	before, err := a.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := a.settings.updateInUserRules(ctx, req.GetEgressProfiles(), req.GetIngressRules())
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, before, "load runtime settings after in-user rule update failed"))
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) changedInUserConnectionUsernamesAfterUpdate(ctx context.Context, before *runtimeSettingsFile, errorMessage string) []string {
	after, err := a.settings.load(ctx)
	if err != nil {
		a.logger.Warn(errorMessage, "error", err)
		return nil
	}
	return changedInUserConnectionUsernames(before, after)
}
