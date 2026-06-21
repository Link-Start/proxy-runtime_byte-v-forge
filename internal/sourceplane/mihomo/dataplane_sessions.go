package mihomo

import (
	"context"
	"sort"

	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
)

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
