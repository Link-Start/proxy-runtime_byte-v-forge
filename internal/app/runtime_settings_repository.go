package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsRepository interface {
	view(context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error)
	load(context.Context) (*runtimeSettingsFile, error)
	update(context.Context, *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateDynamicIPProviders(context.Context, []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateEgressProfiles(context.Context, []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateIngressRules(context.Context, []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	updateInUserRules(context.Context, []*proxyruntimev1.EgressProfileSettings, []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
}

func (a runtimeSettingsApplication) settingsRepository() (runtimeSettingsRepository, error) {
	if a.settings == nil {
		return nil, internalError("runtime settings repository is not configured", nil)
	}
	return a.settings, nil
}

func (a runtimeSettingsApplication) warn(message string, args ...any) {
	if a.logger != nil {
		a.logger.Warn(message, args...)
	}
}
