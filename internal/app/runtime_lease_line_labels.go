package app

import "github.com/byte-v-forge/proxy-runtime/internal/provider"

func applyDynamicLeaseLineLabels(nodes []provider.Node, labels map[string]string) []provider.Node {
	if len(labels) == 0 {
		return nodes
	}
	for index := range nodes {
		nodes[index].Labels = cloneLabels(nodes[index].Labels)
		if nodes[index].Labels == nil {
			nodes[index].Labels = map[string]string{}
		}
		for key, value := range labels {
			nodes[index].Labels[key] = value
		}
	}
	return nodes
}
