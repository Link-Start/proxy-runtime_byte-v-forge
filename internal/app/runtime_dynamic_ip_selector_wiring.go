package app

func runtimeDynamicIPSelectorDependencies(runtime *Runtime) dynamicIPSelectorDependencies {
	if runtime == nil {
		return dynamicIPSelectorDependencies{}
	}
	return dynamicIPSelectorDependencies{
		Store:            runtime.store,
		Settings:         runtime.settings,
		AccountProviders: runtime.accountProviders,
		Concurrency:      runtime.providerConcurrency,
		Logger:           runtime.logger,
		LookupIPGeo:      runtime.lookupIPGeo,
	}
}
