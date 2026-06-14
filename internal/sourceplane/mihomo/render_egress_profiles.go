package mihomo

import "fmt"

func renderEgressProfiles(opts renderOptions) (renderedProfileProjection, error) {
	projection := renderedProfileProjection{
		proxies:   []map[string]any{},
		providers: map[string]mihomoProvider{},
		groups:    []mihomoGroup{},
	}
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
			return renderedProfileProjection{}, fmt.Errorf("profile %q line: %w", id, err)
		}
		projection.addLayer(line)
		exit, err := renderEgressProfileExit(opts, id, profileGroupNameFor(opts, profile), profile.Exit, line.target, opts.BasePool)
		if err != nil {
			return renderedProfileProjection{}, fmt.Errorf("profile %q exit: %w", id, err)
		}
		projection.addLayer(exit)
	}
	return projection, nil
}

func (p *renderedProfileProjection) addLayer(layer renderedProfileLayer) {
	if layer.proxy != nil {
		p.proxies = append(p.proxies, layer.proxy)
	}
	p.proxies = append(p.proxies, layer.proxies...)
	if layer.providerName != "" {
		p.providers[layer.providerName] = layer.provider
	}
	if layer.group.Name != "" {
		p.groups = append(p.groups, layer.group)
	}
}

type renderedProfileLayer struct {
	proxy        map[string]any
	proxies      []map[string]any
	providerName string
	provider     mihomoProvider
	group        mihomoGroup
	target       string
}
