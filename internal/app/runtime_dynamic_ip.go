package app

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/dynamic"
)

func dynamicIPPolicyDurationText(policy *proxygatewayv1.ProxySessionPolicy) string {
	if policy == nil || policy.GetStickyTtl() == nil || policy.GetStickyTtl().AsDuration() <= 0 {
		return ""
	}
	return policy.GetStickyTtl().AsDuration().String()
}

func runtimeDynamicIPSelectorDependencies(runtime *Runtime) dynamic.IPSelectorDependencies {
	if runtime == nil {
		return dynamic.IPSelectorDependencies{}
	}
	return dynamic.IPSelectorDependencies{
		Store:            runtime.store,
		LoadSettings:     runtime.settings.Load,
		AccountProviders: runtime.accountProviders,
		Concurrency:      runtime.providerConcurrency,
		Logger:           runtime.logger,
		LookupIPGeo:      runtime.lookupIPGeo,
		Clock:            runtime.clock,
	}
}
