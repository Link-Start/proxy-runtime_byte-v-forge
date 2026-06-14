package mihomo

type renderedProxyProjection struct {
	proxies         []map[string]any
	providers       map[string]mihomoProvider
	profileGroups   []mihomoGroup
	profileGroupsBy map[string]string
}

func renderProxyProjection(opts renderOptions) (renderedProxyProjection, error) {
	providerMap := cloneNativeProviders(opts.NativeConfig.ProxyProviders)
	fixedConfigs, err := renderBaseProxyConfigs(opts)
	if err != nil {
		return renderedProxyProjection{}, err
	}
	profileGroupsByID := profileGroupNames(opts.EgressProfiles)
	profileProjection, err := renderProfileProjection(opts, fixedConfigs, providerMap, profileGroupsByID)
	if err != nil {
		return renderedProxyProjection{}, err
	}
	fixedConfigs = append(fixedConfigs, profileProjection.proxies...)
	mergeMihomoProviders(providerMap, profileProjection.providers)
	return renderedProxyProjection{
		proxies:         fixedConfigs,
		providers:       providerMap,
		profileGroups:   profileProjection.groups,
		profileGroupsBy: profileGroupsByID,
	}, nil
}
