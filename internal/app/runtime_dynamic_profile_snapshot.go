package app

import (
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func (r *Runtime) setDynamicProfilePoolSnapshot(nodes []provider.Node) {
	r.dynamicProfileMu.Lock()
	defer r.dynamicProfileMu.Unlock()
	r.dynamicProfilePoolNodes = cloneProviderNodes(nodes)
	r.dynamicProfileUpdatedAt = r.clock.Now().UTC()
}

func (r *Runtime) dynamicProfilePoolSnapshot() ([]provider.Node, time.Time) {
	r.dynamicProfileMu.RLock()
	defer r.dynamicProfileMu.RUnlock()
	return cloneProviderNodes(r.dynamicProfilePoolNodes), r.dynamicProfileUpdatedAt
}

func cloneProviderNodes(nodes []provider.Node) []provider.Node {
	out := make([]provider.Node, 0, len(nodes))
	for _, node := range nodes {
		item := node
		if node.URL != nil {
			copied := *node.URL
			item.URL = &copied
		}
		item.Labels = cloneLabels(node.Labels)
		out = append(out, item)
	}
	return out
}
