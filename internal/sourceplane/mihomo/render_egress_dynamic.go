package mihomo

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func renderDynamicProfileExit(profileID string, groupName string, exit sourceplane.EgressProfileExit, dialerProxy string, nodes []provider.Node) (renderedProfileLayer, error) {
	group := profileLayerGroup(groupName, exit, "select")
	nodes = dynamicProfileNodes(nodes, profileID, exit.ProviderID)
	if len(nodes) == 0 {
		group.Proxies = []string{"REJECT"}
		return renderedProfileLayer{group: group}, nil
	}
	if strings.TrimSpace(dialerProxy) == "" {
		group.Proxies = make([]string, 0, len(nodes))
		for index, node := range nodes {
			group.Proxies = append(group.Proxies, providerNodeName("provider-pool", node, index))
		}
		return renderedProfileLayer{group: group}, nil
	}
	out := renderedProfileLayer{group: group, proxies: make([]map[string]any, 0, len(nodes))}
	for index, node := range nodes {
		name := profileDynamicProxyName(profileID, index, node)
		proxy, err := renderProxyURL(name, node.URL)
		if err != nil {
			return renderedProfileLayer{}, err
		}
		proxy["dialer-proxy"] = strings.TrimSpace(dialerProxy)
		out.proxies = append(out.proxies, proxy)
		out.group.Proxies = append(out.group.Proxies, name)
	}
	return out, nil
}
