package mihomo

import (
	"fmt"

	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

func renderProviderNodes(prefix string, nodes []provider.Node) ([]map[string]any, []string, error) {
	proxies := make([]map[string]any, 0, len(nodes))
	names := make([]string, 0, len(nodes))
	for index, node := range nodes {
		name := providerNodeName(prefix, node, index)
		proxy, err := renderProxyURL(name, node.URL)
		if err != nil {
			return nil, nil, fmt.Errorf("render proxy node %q: %w", name, err)
		}
		proxies = append(proxies, proxy)
		names = append(names, name)
	}
	return proxies, names, nil
}

func providerNodeName(prefix string, node provider.Node, index int) string {
	name := safeID(firstNonEmpty(node.ID, fmt.Sprintf("%s-%d", prefix, index)))
	if name == "" {
		return fmt.Sprintf("%s-%d", prefix, index)
	}
	return name
}
