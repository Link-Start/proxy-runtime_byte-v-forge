package mihomo

import (
	"fmt"
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

func dynamicProfileNodes(nodes []provider.Node, profileID string, dynamicProviderID string) []provider.Node {
	profileID = safeID(profileID)
	dynamicProviderID = strings.TrimSpace(dynamicProviderID)
	out := make([]provider.Node, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.Labels["egress_profile_id"]) != profileID {
			continue
		}
		if dynamicProviderID == "" {
			out = append(out, node)
			continue
		}
		if dynamicProfileNodeHasProviderID(node, dynamicProviderID) {
			out = append(out, node)
		}
	}
	return out
}

func dynamicProfileNodeHasProviderID(node provider.Node, dynamicProviderID string) bool {
	dynamicProviderID = strings.TrimSpace(dynamicProviderID)
	if strings.TrimSpace(node.Labels["dynamic_provider_id"]) == dynamicProviderID || strings.TrimSpace(node.ProviderID) == dynamicProviderID {
		return true
	}
	for _, value := range strings.Split(node.Labels["dynamic_provider_ids"], ",") {
		if strings.TrimSpace(value) == dynamicProviderID {
			return true
		}
	}
	return false
}

func profileDynamicProxyName(profileID string, index int, node provider.Node) string {
	return safeID(fmt.Sprintf("%s-dynamic-%d-%s", profileInternalGroupName(profileID), index, providerNodeName("provider-pool", node, index)))
}
