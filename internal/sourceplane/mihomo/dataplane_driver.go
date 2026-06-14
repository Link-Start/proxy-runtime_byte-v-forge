package mihomo

import (
	"context"
	"sort"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) ApplyDesiredConfig(ctx context.Context, cfg dataplane.Config) ([]provider.Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	previous := cloneDataPlaneConfig(d.baseCfg)
	d.baseCfg = cloneDataPlaneConfig(cfg)
	nodes, err := d.reconcileLocked(ctx, sourceConfigFromDataPlane(d.baseCfg))
	if err != nil {
		d.baseCfg = previous
		return nil, err
	}
	return nodes, nil
}

func (d *Driver) UpsertSessionRoute(ctx context.Context, route dataplane.SessionRoute) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	key := sessionRouteKey(route)
	previous, existed := d.sessions[key]
	d.sessions[key] = cloneSessionRoute(route)
	_, err := d.reconcileLocked(ctx, sourceConfigFromDataPlane(d.baseCfg))
	if err != nil {
		if existed {
			d.sessions[key] = previous
		} else {
			delete(d.sessions, key)
		}
	}
	return err
}

func (d *Driver) DeleteSessionRoute(ctx context.Context, route dataplane.SessionRoute) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	key := sessionRouteKey(route)
	previous, existed := d.sessions[key]
	delete(d.sessions, key)
	_, err := d.reconcileLocked(ctx, sourceConfigFromDataPlane(d.baseCfg))
	if err != nil && existed {
		d.sessions[key] = previous
	}
	return err
}

func (d *Driver) sessionRoutesLocked() []dataplane.SessionRoute {
	keys := make([]string, 0, len(d.sessions))
	for key := range d.sessions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]dataplane.SessionRoute, 0, len(keys))
	for _, key := range keys {
		out = append(out, cloneSessionRoute(d.sessions[key]))
	}
	return out
}

func sessionRouteKey(route dataplane.SessionRoute) string {
	return firstNonEmpty(route.SessionID, route.Listener.Name)
}

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
