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
