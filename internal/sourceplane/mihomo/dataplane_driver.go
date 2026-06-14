package mihomo

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
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
