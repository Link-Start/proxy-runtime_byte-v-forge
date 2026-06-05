package mihomo

import (
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func renderConfig(opts renderOptions) (mihomoConfig, error) {
	providerMap := make(map[string]mihomoProvider, len(opts.Providers))
	providerIDs := make([]string, 0, len(opts.Providers))
	for _, item := range opts.Providers {
		id := safeID(item.ID)
		providerIDs = append(providerIDs, id)
		providerMap[id] = renderSubscriptionProvider(opts, item, id, "")
	}
	fixedConfigs := make([]map[string]any, 0, len(opts.FixedProxies))
	fixedIDs := make([]string, 0, len(opts.FixedProxies))
	for _, item := range opts.FixedProxies {
		proxyConfig, err := renderFixedProxy(item)
		if err != nil {
			return mihomoConfig{}, err
		}
		fixedConfigs = append(fixedConfigs, proxyConfig)
		fixedIDs = append(fixedIDs, safeID(item.ID))
	}
	poolConfigs, poolIDs, err := renderProviderNodes("provider-pool", opts.BasePool)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, poolConfigs...)
	fixedIDs = append(fixedIDs, poolIDs...)
	defaultProxies := append([]string(nil), fixedIDs...)
	if len(defaultProxies) == 0 && len(providerIDs) == 0 {
		defaultProxies = []string{"DIRECT"}
	}
	sessionConfigs, err := renderSessionRoutes(opts.SessionRoutes)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, sessionConfigs...)
	profileConfigs, profileProviders, profileGroups, err := renderEgressProfiles(opts)
	if err != nil {
		return mihomoConfig{}, err
	}
	fixedConfigs = append(fixedConfigs, profileConfigs...)
	for id, provider := range profileProviders {
		providerMap[id] = provider
	}
	gateway, userGroups, userRules, err := renderGateway(opts.Endpoint, opts.ProxyUsers, opts.SessionRoutes)
	if err != nil {
		return mihomoConfig{}, err
	}
	rules := append(userRules, "MATCH,"+groupName)
	return mihomoConfig{
		MixedPort:          gateway.Port,
		BindAddress:        gateway.Listen,
		AllowLAN:           true,
		Mode:               "rule",
		LogLevel:           "warning",
		ExternalController: strings.TrimSpace(opts.APIAddr),
		ExternalUI:         strings.TrimSpace(opts.DashboardDir),
		ExternalUIURL:      strings.TrimSpace(opts.DashboardURL),
		Authentication:     renderAuthentication(gateway.Users),
		Proxies:            fixedConfigs,
		ProxyProviders:     providerMap,
		ProxyGroups: append([]mihomoGroup{{
			Name:           groupName,
			Type:           groupStrategy(opts.GroupStrategy),
			Proxies:        defaultProxies,
			Use:            providerIDs,
			URL:            firstNonEmpty(opts.HealthCheckURL, "https://www.gstatic.com/generate_204"),
			Interval:       secondsDuration(opts.HealthCheckInterval, 300),
			Timeout:        millisecondsDuration(opts.HealthCheckTimeout, 5000),
			Lazy:           true,
			ExpectedStatus: 204,
		}}, append(profileGroups, userGroups...)...),
		Rules: rules,
	}, nil
}

func renderSubscriptionProvider(opts renderOptions, item sourceplane.SubscriptionProvider, id string, dialerProxy string) mihomoProvider {
	provider := mihomoProvider{
		Type:     "http",
		URL:      item.URL,
		Path:     providerConfigPath(opts.ConfigDir, item, id),
		Interval: seconds(item.Interval, 3600),
		Filter:   item.Filter,
		Exclude:  item.ExcludeFilter,
		Header:   item.Headers,
		HealthCheck: &mihomoHealthCheck{
			Enable:         true,
			URL:            firstNonEmpty(item.HealthCheckURL, opts.HealthCheckURL, "https://www.gstatic.com/generate_204"),
			Interval:       seconds(item.HealthInterval, secondsDuration(opts.HealthCheckInterval, 300)),
			Timeout:        milliseconds(item.HealthTimeout, millisecondsDuration(opts.HealthCheckTimeout, 5000)),
			Lazy:           item.HealthLazy,
			ExpectedStatus: defaultExpectedStatus(item.ExpectedStatus),
		},
	}
	if strings.TrimSpace(dialerProxy) != "" {
		provider.Override = map[string]any{
			"additional-prefix": safeID(id) + "-",
			"dialer-proxy":      strings.TrimSpace(dialerProxy),
		}
	}
	return provider
}

func groupStrategy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "select", "url-test", "load-balance":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "fallback"
	}
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
