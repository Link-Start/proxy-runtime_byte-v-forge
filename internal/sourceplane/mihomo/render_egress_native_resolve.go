package mihomo

import "strings"

func mihomoNativeNodeName(resourceID string, nodeID string) string {
	resourceID = strings.TrimSpace(resourceID)
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return ""
	}
	prefix := resourceID + "/"
	if strings.HasPrefix(nodeID, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(nodeID, prefix))
	}
	return nodeID
}

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

func mihomoNativeProviderNodeName(resourceID string, providerName string, nodeID string) string {
	for _, prefix := range []string{strings.TrimSpace(resourceID), strings.TrimSpace(providerName)} {
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(nodeID, prefix+"/") {
			return strings.TrimSpace(strings.TrimPrefix(nodeID, prefix+"/"))
		}
	}
	return strings.TrimSpace(nodeID)
}

func mihomoProxyAvailable(available map[string]struct{}, name string) bool {
	if len(available) == 0 {
		return false
	}
	_, exists := available[strings.TrimSpace(name)]
	return exists
}

func mihomoProviderAvailable(available map[string]struct{}, name string) bool {
	if len(available) == 0 {
		return false
	}
	_, exists := available[strings.TrimSpace(name)]
	return exists
}
