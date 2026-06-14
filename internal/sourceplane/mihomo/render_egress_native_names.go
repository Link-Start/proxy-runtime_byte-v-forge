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
