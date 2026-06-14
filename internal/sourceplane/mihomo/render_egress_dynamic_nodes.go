package mihomo

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

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
