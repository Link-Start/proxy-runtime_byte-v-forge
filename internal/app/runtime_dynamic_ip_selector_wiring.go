package app

import "github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"

func runtimeDynamicIPSelectorDependencies(runtime *Runtime) dynamic.IPSelectorDependencies {
	if runtime == nil {
		return dynamic.IPSelectorDependencies{}
	}
	return dynamic.IPSelectorDependencies{
		Store:            runtime.store,
		LoadSettings:     runtime.settings.load,
		AccountProviders: runtime.accountProviders,
		Concurrency:      runtime.providerConcurrency,
		Logger:           runtime.logger,
		LookupIPGeo:      runtime.lookupIPGeo,
		Clock:            runtime.clock,
	}
}
