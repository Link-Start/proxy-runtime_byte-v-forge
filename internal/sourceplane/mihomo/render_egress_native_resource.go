package mihomo

import "strings"

type resolvedMihomoNativeResource struct {
	ResourceName string
	NodeName     string
}

func resolveMihomoNativeResource(config mihomoNativeConfig, resourceID string, nodeID string) resolvedMihomoNativeResource {
	resourceID = strings.TrimSpace(resourceID)
	nodeID = strings.TrimSpace(nodeID)
	for _, proxy := range config.FixedProxies {
		id := strings.TrimSpace(proxy.ID)
		name := strings.TrimSpace(proxy.Name)
		if name == "" {
			continue
		}
		if resourceID == id || resourceID == name {
			return resolvedMihomoNativeResource{ResourceName: name, NodeName: name}
		}
	}
	for _, subscription := range config.Subscriptions {
		id := strings.TrimSpace(subscription.ID)
		name := strings.TrimSpace(subscription.Name)
		if name == "" {
			continue
		}
		if resourceID == id || resourceID == name {
			return resolvedMihomoNativeResource{ResourceName: name, NodeName: mihomoNativeProviderNodeName(id, name, nodeID)}
		}
	}
	return resolvedMihomoNativeResource{}
}
