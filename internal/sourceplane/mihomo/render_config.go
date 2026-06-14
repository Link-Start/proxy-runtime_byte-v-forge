package mihomo

func renderConfig(opts renderOptions) (mihomoConfig, error) {
	proxyProjection, err := renderProxyProjection(opts)
	if err != nil {
		return mihomoConfig{}, err
	}
	gatewayProjection, err := renderGatewayProjection(opts, proxyProjection.profileGroupsBy)
	if err != nil {
		return mihomoConfig{}, err
	}
	groups := appendUniqueGroups(baseMihomoGroups(), opts.NativeConfig.ProxyGroups, proxyProjection.profileGroups, gatewayProjection.groups)
	config := newBaseRenderedConfig(opts, gatewayProjection.listener)
	config.Proxies = proxyProjection.proxies
	config.ProxyProviders = proxyProjection.providers
	config.ProxyGroups = groups
	config.Rules = renderConfigRules(gatewayProjection.rules, opts.NativeConfig.Rules)
	return config, nil
}
