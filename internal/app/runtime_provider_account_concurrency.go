package app

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"
	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

const providerAccountConcurrencyTTLBuffer = 2 * time.Minute

func dynamicProviderConcurrencyLimit(settings *runtimeSettingsFile, dynamicProviderID string, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	dynamicProviderID = appcore.RuntimeSafeID(dynamicProviderID)
	for _, provider := range dynamic.ProviderInstances(settings) {
		if provider.DynamicProviderID == dynamicProviderID {
			return dynamic.ProviderInstanceConcurrencyLimit(provider, policy)
		}
	}
	if leaseapp.ConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return settingscore.DefaultDynamicProviderRotatingConcurrencyLimit
	}
	return settingscore.DefaultDynamicProviderStickyConcurrencyLimit
}
