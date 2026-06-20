package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type leaseSettingsAdapterFactory struct {
	deps leaseCoordinatorDependencies
}

func (f leaseSettingsAdapterFactory) New() leaseapp.SettingsAdapter[*runtimeSettingsFile] {
	return leaseapp.SettingsAdapter[*runtimeSettingsFile]{
		Load:                    f.deps.settings.Load,
		ResolveProviderGateways: leaseProviderGateways,
		ResolveLineBinding:      f.resolveLineBinding,
	}
}

func (f leaseSettingsAdapterFactory) resolveLineBinding(ctx context.Context, settings *runtimeSettingsFile, accountID string) (string, map[string]string, error) {
	return f.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
}

func (c leaseCoordinator) settingsAdapter() leaseapp.SettingsAdapter[*runtimeSettingsFile] {
	return leaseSettingsAdapterFactory{deps: c.deps}.New()
}

func leaseProviderGateways(settings *runtimeSettingsFile, lease *proxyruntimev1.ProxyDynamicLease, providerID string) ([]accountproxy.Gateway, error) {
	return dynamic.EndpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerID), nil
}
