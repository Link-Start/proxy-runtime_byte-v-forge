package mihomo

import "github.com/byte-v-forge/proxy-gateway/internal/provider"

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
