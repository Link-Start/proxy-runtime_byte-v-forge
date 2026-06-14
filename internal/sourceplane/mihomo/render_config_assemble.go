package mihomo

func assembleRenderedConfig(opts renderOptions, sections renderedConfigSections) mihomoConfig {
	config := newBaseRenderedConfig(opts, sections.listener)
	config.Proxies = sections.proxies
	config.ProxyProviders = sections.providers
	config.ProxyGroups = sections.groups
	config.Rules = sections.rules
	return config
}
