package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"
)

func dynamicIPPolicyDurationText(policy *proxyruntimev1.ProxySessionPolicy) string {
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
