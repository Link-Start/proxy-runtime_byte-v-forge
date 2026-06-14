package settings

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type Repository interface {
	View(context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error)
	Load(context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error)
	Update(context.Context, *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error)
	UpdateDynamicIPProviders(context.Context, []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	UpdateEgressProfiles(context.Context, []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	UpdateIngressRules(context.Context, []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
	UpdateInUserRules(context.Context, []*proxyruntimev1.EgressProfileSettings, []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error)
}
