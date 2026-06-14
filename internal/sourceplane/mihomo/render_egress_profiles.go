package mihomo

import "fmt"

func renderEgressProfiles(opts renderOptions) ([]map[string]any, map[string]mihomoProvider, []mihomoGroup, error) {
	proxies := []map[string]any{}
	providers := map[string]mihomoProvider{}
	groups := []mihomoGroup{}
	for _, profile := range opts.EgressProfiles {
		if !profile.Enabled {
			continue
		}
		id := safeID(profile.ID)
		if id == "" {
			continue
		}
		line, err := renderEgressProfileLine(opts, id, profile.Line)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("profile %q line: %w", id, err)
		}
		if line.proxy != nil {
			proxies = append(proxies, line.proxy)
		}
		if line.providerName != "" {
			providers[line.providerName] = line.provider
		}
		if line.group.Name != "" {
			groups = append(groups, line.group)
		}
		exit, err := renderEgressProfileExit(opts, id, profileGroupNameFor(opts, profile), profile.Exit, line.target, opts.BasePool)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("profile %q exit: %w", id, err)
		}
		if exit.proxy != nil {
			proxies = append(proxies, exit.proxy)
		}
		proxies = append(proxies, exit.proxies...)
		if exit.providerName != "" {
			providers[exit.providerName] = exit.provider
		}
		groups = append(groups, exit.group)
	}
	return proxies, providers, groups, nil
}

type renderedProfileLayer struct {
	proxy        map[string]any
	proxies      []map[string]any
	providerName string
	provider     mihomoProvider
	group        mihomoGroup
	target       string
}
