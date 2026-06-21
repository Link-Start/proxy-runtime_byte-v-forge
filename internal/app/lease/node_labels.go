package lease

import (
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

func ApplyNodeLabels(nodes []provider.Node, labels map[string]string) []provider.Node {
	if len(labels) == 0 {
		return nodes
	}
	for index := range nodes {
		nodes[index].Labels = appcore.CloneStringMap(nodes[index].Labels)
		if nodes[index].Labels == nil {
			nodes[index].Labels = map[string]string{}
		}
		for key, value := range labels {
			nodes[index].Labels[key] = value
		}
	}
	return nodes
}
