package app

import (
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/dynamic"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

const providerAccountConcurrencyTTLBuffer = 2 * time.Minute

func dynamicProviderConcurrencyLimit(settings *runtimeSettingsFile, dynamicProviderID string, policy *proxygatewayv1.ProxySessionPolicy) uint32 {
	dynamicProviderID = appcore.RuntimeSafeID(dynamicProviderID)
	for _, provider := range dynamic.ProviderInstances(settings) {
		if provider.DynamicProviderID == dynamicProviderID {
			return dynamic.ProviderInstanceConcurrencyLimit(provider, policy)
		}
	}
	if leaseapp.ConcurrencyMode(policy) == proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return kernel.DefaultDynamicProviderRotatingConcurrencyLimit
	}
	return kernel.DefaultDynamicProviderStickyConcurrencyLimit
}
