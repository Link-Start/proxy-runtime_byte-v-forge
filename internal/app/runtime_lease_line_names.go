package app

import "strings"

func mihomoNodeResourcePrefix(nodeID string) string {
	resource, _, ok := strings.Cut(strings.TrimSpace(nodeID), "/")
	if !ok {
		return ""
	}
	return strings.TrimSpace(resource)
}

func dynamicLeaseProfileLineGroupName(profileID string) string {
	id := runtimeSafeID(profileID)
	if id == "" {
		id = "profile"
	}
	return "bvf-profile-" + id + "-line"
}

func mihomoNodeDialerProxyName(resourceID string, nodeID string) string {
	resourceID = strings.TrimSpace(resourceID)
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return ""
	}
	prefix := resourceID + "/"
	if resourceID != "" && strings.HasPrefix(nodeID, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(nodeID, prefix))
	}
	return nodeID
}
