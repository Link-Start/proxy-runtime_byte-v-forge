package mihomo

import (
	"strings"
	"time"
)

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

func defaultExpectedStatus(value uint32) uint32 {
	if value == 0 {
		return 204
	}
	return value
}

func seconds(value time.Duration, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return int(value / time.Second)
}

func milliseconds(value time.Duration, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return int(value / time.Millisecond)
}

func secondsDuration(value time.Duration, fallback int) int { return seconds(value, fallback) }

func millisecondsDuration(value time.Duration, fallback int) int {
	return milliseconds(value, fallback)
}

func mihomoProxyNames(proxies []map[string]any) map[string]struct{} {
	out := map[string]struct{}{}
	for _, proxy := range proxies {
		name, _ := proxy["name"].(string)
		if name = strings.TrimSpace(name); name != "" {
			out[name] = struct{}{}
		}
	}
	return out
}

func mihomoProviderNames(providers map[string]mihomoProvider) map[string]struct{} {
	out := map[string]struct{}{}
	for name := range providers {
		if name = strings.TrimSpace(name); name != "" {
			out[name] = struct{}{}
		}
	}
	return out
}

func appendUniqueGroups(base []mihomoGroup, groups ...[]mihomoGroup) []mihomoGroup {
	out := make([]mihomoGroup, 0, len(base))
	seen := map[string]struct{}{}
	for _, group := range base {
		if group.Name == "" {
			continue
		}
		seen[group.Name] = struct{}{}
		out = append(out, group)
	}
	for _, items := range groups {
		for _, group := range items {
			if group.Name == "" {
				continue
			}
			if _, exists := seen[group.Name]; exists {
				continue
			}
			seen[group.Name] = struct{}{}
			out = append(out, group)
		}
	}
	return out
}
