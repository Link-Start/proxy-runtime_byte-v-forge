package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) settingsAdapter() leaseapp.SettingsAdapter[*runtimeSettingsFile] {
	return leaseapp.SettingsAdapter[*runtimeSettingsFile]{
		Load: c.deps.settings.load,
		ResolveProviderGateways: func(settings *runtimeSettingsFile, lease *proxyruntimev1.ProxyDynamicLease, providerID string) ([]accountproxy.Gateway, error) {
			return endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerID), nil
		},
		ResolveLineBinding: func(ctx context.Context, settings *runtimeSettingsFile, accountID string) (string, map[string]string, error) {
			return c.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
		},
	}
}
