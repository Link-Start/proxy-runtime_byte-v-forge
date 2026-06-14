package mihomo

type renderedProfileProjection struct {
	proxies   []map[string]any
	providers map[string]mihomoProvider
	groups    []mihomoGroup
}

func renderProfileProjection(opts renderOptions, fixedConfigs []map[string]any, providerMap map[string]mihomoProvider, profileGroupsByID map[string]string) (renderedProfileProjection, error) {
	profileOpts := opts
	profileOpts.AvailableProxies = mihomoProxyNames(fixedConfigs)
	profileOpts.AvailableProviders = mihomoProviderNames(providerMap)
	profileOpts.ProfileGroups = profileGroupsByID
	profileConfigs, profileProviders, profileGroups, err := renderEgressProfiles(profileOpts)
	if err != nil {
		return renderedProfileProjection{}, err
	}
	return renderedProfileProjection{proxies: profileConfigs, providers: profileProviders, groups: profileGroups}, nil
}

func mergeMihomoProviders(target map[string]mihomoProvider, providers map[string]mihomoProvider) {
	for id, provider := range providers {
		target[id] = provider
	}
}
