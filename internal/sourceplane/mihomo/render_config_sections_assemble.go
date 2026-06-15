package mihomo

func assembleRenderedConfigSections(opts renderOptions, parts renderedConfigProjectionParts) renderedConfigSections {
	return renderedConfigSections{
		listener:  parts.gateway.listener,
		proxies:   parts.proxy.proxies,
		providers: parts.proxy.providers,
		groups:    renderConfigGroups(opts, parts),
		rules:     renderConfigRules(parts.gateway.rules, opts.NativeConfig.Rules),
	}
}

func renderConfigGroups(opts renderOptions, parts renderedConfigProjectionParts) []mihomoGroup {
	return appendUniqueGroups(
		baseMihomoGroups(),
		opts.NativeConfig.ProxyGroups,
		parts.proxy.profileGroups,
		parts.gateway.groups,
	)
}
