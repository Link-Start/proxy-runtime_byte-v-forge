package mihomo

type renderedProxyProjection struct {
	proxies         []map[string]any
	providers       map[string]mihomoProvider
	profileGroups   []mihomoGroup
	profileGroupsBy map[string]string
}

type renderedGatewayProjection struct {
	listener mihomoListener
	groups   []mihomoGroup
	rules    []string
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

func renderGatewayProjection(opts renderOptions, profileGroupsByID map[string]string) (renderedGatewayProjection, error) {
	gateway, userGroups, userRules, err := renderGateway(opts.Endpoint, opts.ProxyUsers, opts.SessionRoutes, profileGroupsByID)
	if err != nil {
		return renderedGatewayProjection{}, err
	}
	return renderedGatewayProjection{listener: gateway, groups: userGroups, rules: userRules}, nil
}
