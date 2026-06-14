package mihomo

import "strings"

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
	return mihomoConfig{
		MixedPort:          gatewayProjection.listener.Port,
		BindAddress:        gatewayProjection.listener.Listen,
		AllowLAN:           true,
		Mode:               "rule",
		LogLevel:           "warning",
		ExternalController: strings.TrimSpace(opts.APIAddr),
		Secret:             strings.TrimSpace(opts.ControllerSecret),
		ExternalUI:         strings.TrimSpace(opts.DashboardDir),
		ExternalUIURL:      strings.TrimSpace(opts.DashboardURL),
		Authentication:     renderAuthentication(gatewayProjection.listener.Users),
		Proxies:            proxyProjection.proxies,
		ProxyProviders:     proxyProjection.providers,
		ProxyGroups:        groups,
		Rules:              renderConfigRules(gatewayProjection.rules, opts.NativeConfig.Rules),
	}, nil
}
