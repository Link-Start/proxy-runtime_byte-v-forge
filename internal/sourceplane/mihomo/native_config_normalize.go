package mihomo

func normalizeNativeConfigPaths(configDir string, config *mihomoNativeConfig) {
	if config == nil || configDir == "" {
		return
	}
	for name, provider := range config.ProxyProviders {
		provider.Path = normalizeNativeProviderPath(configDir, provider.Path)
		normalizeNativeProviderHeaders(&provider)
		config.ProxyProviders[name] = provider
	}
	normalizeNativeGroups(config)
}
