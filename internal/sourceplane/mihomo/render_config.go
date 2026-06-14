package mihomo

import "strings"

func renderConfig(opts renderOptions) (mihomoConfig, error) {
	providerMap := cloneNativeProviders(opts.NativeConfig.ProxyProviders)
	fixedConfigs, err := renderBaseProxyConfigs(opts)
	if err != nil {
		return mihomoConfig{}, err
	}
	profileGroupsByID := profileGroupNames(opts.EgressProfiles)
	profileProjection, err := renderProfileProjection(opts, fixedConfigs, providerMap, profileGroupsByID)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, profileProjection.proxies...)
	mergeMihomoProviders(providerMap, profileProjection.providers)
	gateway, userGroups, userRules, err := renderGateway(opts.Endpoint, opts.ProxyUsers, opts.SessionRoutes, profileGroupsByID)
	if err != nil {
		return mihomoConfig{}, err
	}
	groups := appendUniqueGroups(baseMihomoGroups(), opts.NativeConfig.ProxyGroups, profileProjection.groups, userGroups)
	return mihomoConfig{
		MixedPort:          gateway.Port,
		BindAddress:        gateway.Listen,
		AllowLAN:           true,
		Mode:               "rule",
		LogLevel:           "warning",
		ExternalController: strings.TrimSpace(opts.APIAddr),
		Secret:             strings.TrimSpace(opts.ControllerSecret),
		ExternalUI:         strings.TrimSpace(opts.DashboardDir),
		ExternalUIURL:      strings.TrimSpace(opts.DashboardURL),
		Authentication:     renderAuthentication(gateway.Users),
		Proxies:            fixedConfigs,
		ProxyProviders:     providerMap,
		ProxyGroups:        groups,
		Rules:              renderConfigRules(userRules, opts.NativeConfig.Rules),
	}, nil
}
