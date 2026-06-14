package lease

import "github.com/byte-v-forge/proxy-runtime/internal/provider"

func ApplyNodeLabels(nodes []provider.Node, labels map[string]string) []provider.Node {
	if len(labels) == 0 {
		return nodes
	}
	for index := range nodes {
		nodes[index].Labels = cloneStringMap(nodes[index].Labels)
		if nodes[index].Labels == nil {
			nodes[index].Labels = map[string]string{}
		}
		for key, value := range labels {
			nodes[index].Labels[key] = value
		}
	}
	return nodes
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
