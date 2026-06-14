package mihomo

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
