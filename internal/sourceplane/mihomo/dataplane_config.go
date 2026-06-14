package mihomo

import (
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func sourceConfigFromDataPlane(cfg dataplane.Config) sourceplane.Config {
	return sourceplane.Config{
		EgressProfiles:      cfg.EgressProfiles,
		Endpoint:            cfg.Endpoint,
		HealthCheckURL:      cfg.HealthCheckURL,
		HealthCheckInterval: cfg.HealthCheckPeriod,
		HealthCheckTimeout:  cfg.HealthCheckWait,
	}
}

func cloneDataPlaneConfig(cfg dataplane.Config) dataplane.Config {
	cfg.EgressProfiles = cloneEgressProfiles(cfg.EgressProfiles)
	cfg.Pool = cloneProviderNodes(cfg.Pool)
	cfg.ProxyUsers = append([]dataplane.ProxyUserRoute(nil), cfg.ProxyUsers...)
	return cfg
}

func cloneEgressProfiles(profiles []sourceplane.EgressProfile) []sourceplane.EgressProfile {
	if len(profiles) == 0 {
		return nil
	}
	out := make([]sourceplane.EgressProfile, len(profiles))
	for index, profile := range profiles {
		out[index] = profile
	}
	return out
}

func cloneSessionRoute(route dataplane.SessionRoute) dataplane.SessionRoute {
	route.Pool = cloneProviderNodes(route.Pool)
	return route
}

func cloneProviderNodes(nodes []provider.Node) []provider.Node {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]provider.Node, 0, len(nodes))
	for _, node := range nodes {
		copy := node
		if node.URL != nil {
			cloned := *node.URL
			copy.URL = &cloned
		}
		if len(node.Labels) > 0 {
			copy.Labels = make(map[string]string, len(node.Labels))
			for key, value := range node.Labels {
				copy.Labels[key] = value
			}
		}
		out = append(out, copy)
	}
	return out
}
