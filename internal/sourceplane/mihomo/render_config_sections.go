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

type renderedConfigSections struct {
	listener  mihomoListener
	proxies   []map[string]any
	providers map[string]mihomoProvider
	groups    []mihomoGroup
	rules     []string
}

func renderConfigSections(opts renderOptions) (renderedConfigSections, error) {
	proxyProjection, err := renderProxyProjection(opts)
	if err != nil {
		return renderedConfigSections{}, err
	}
	gatewayProjection, err := renderGatewayProjection(opts, proxyProjection.profileGroupsBy)
	if err != nil {
		return renderedConfigSections{}, err
	}
	return renderedConfigSections{
		listener:  gatewayProjection.listener,
		proxies:   proxyProjection.proxies,
		providers: proxyProjection.providers,
		groups:    appendUniqueGroups(baseMihomoGroups(), opts.NativeConfig.ProxyGroups, proxyProjection.profileGroups, gatewayProjection.groups),
		rules:     renderConfigRules(gatewayProjection.rules, opts.NativeConfig.Rules),
	}, nil
}

func assembleRenderedConfig(opts renderOptions, sections renderedConfigSections) mihomoConfig {
	config := newBaseRenderedConfig(opts, sections.listener)
	config.Proxies = sections.proxies
	config.ProxyProviders = sections.providers
	config.ProxyGroups = sections.groups
	config.Rules = sections.rules
	return config
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
	gateway, err := renderGateway(opts.Endpoint, opts.ProxyUsers, opts.SessionRoutes, profileGroupsByID)
	if err != nil {
		return renderedGatewayProjection{}, err
	}
	return renderedGatewayProjection{listener: gateway.listener, groups: gateway.groups, rules: gateway.rules}, nil
}
