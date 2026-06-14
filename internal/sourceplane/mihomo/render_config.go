package mihomo

import "strings"

func renderConfig(opts renderOptions) (mihomoConfig, error) {
	providerMap := cloneNativeProviders(opts.NativeConfig.ProxyProviders)
	fixedConfigs := cloneNativeProxies(opts.NativeConfig.Proxies)
	poolConfigs, _, err := renderProviderNodes("provider-pool", opts.BasePool)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, poolConfigs...)
	sessionConfigs, err := renderSessionRoutes(opts.SessionRoutes)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, sessionConfigs...)
	profileGroupsByID := profileGroupNames(opts.EgressProfiles)
	profileOpts := opts
	profileOpts.AvailableProxies = mihomoProxyNames(fixedConfigs)
	profileOpts.AvailableProviders = mihomoProviderNames(providerMap)
	profileOpts.ProfileGroups = profileGroupsByID
	profileConfigs, profileProviders, profileGroups, err := renderEgressProfiles(profileOpts)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, profileConfigs...)
	for id, provider := range profileProviders {
		providerMap[id] = provider
	}
	gateway, userGroups, userRules, err := renderGateway(opts.Endpoint, opts.ProxyUsers, opts.SessionRoutes, profileGroupsByID)
	if err != nil {
		return mihomoConfig{}, err
	}
	rules := append(userRules, opts.NativeConfig.Rules...)
	rules = append(rules, "MATCH,REJECT")
	baseGroups := []mihomoGroup{
		{
			Name:    "GLOBAL",
			Type:    "select",
			Proxies: []string{"REJECT"},
			Hidden:  true,
		},
	}
	groups := appendUniqueGroups(baseGroups, opts.NativeConfig.ProxyGroups, profileGroups, userGroups)
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
		Rules:              rules,
	}, nil
}
