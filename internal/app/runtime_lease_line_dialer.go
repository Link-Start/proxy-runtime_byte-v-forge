package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func dynamicLeaseProfileDialerProxy(profile *proxyruntimev1.EgressProfileSettings, nativeConfig mihomoNativeConfigFile) (string, map[string]string, error) {
	node := profile.GetLine().GetMihomoNode()
	dialer := dynamicLeaseLineDialerProxy(profile.GetProfileId(), nativeConfig, node.GetResourceId(), node.GetNodeId())
	if dialer == "" {
		return "", nil, fmt.Errorf(
			"egress profile %q line mihomo node %q/%q cannot be resolved from current native config",
			strings.TrimSpace(profile.GetProfileId()),
			strings.TrimSpace(node.GetResourceId()),
			strings.TrimSpace(node.GetNodeId()),
		)
	}
	return dialer, map[string]string{
		"line_kind":        "mihomo_node",
		"line_resource_id": strings.TrimSpace(node.GetResourceId()),
		"line_node_id":     strings.TrimSpace(node.GetNodeId()),
		"line_dialer":      dialer,
	}, nil
}

func dynamicLeaseLineDialerProxy(profileID string, nativeConfig mihomoNativeConfigFile, resourceID string, nodeID string) string {
	resourceID = strings.TrimSpace(resourceID)
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return ""
	}
	if dialer := dynamicLeaseNativeDialerProxy(profileID, nativeConfig, resourceID, nodeID); dialer != "" {
		return dialer
	}
	if resourceID == "" {
		return strings.TrimSpace(nodeID)
	}
	return ""
}

func dynamicLeaseNativeDialerProxy(profileID string, nativeConfig mihomoNativeConfigFile, resourceID string, nodeID string) string {
	fixedByID, fixedByName := currentFixedProxyIndexes(nativeConfig)
	for _, key := range []string{resourceID, nodeID, mihomoNodeDialerProxyName(resourceID, nodeID)} {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if proxy := fixedByID[key]; proxy.Name != "" {
			return proxy.Name
		}
		if proxy := fixedByName[key]; proxy.Name != "" {
			return proxy.Name
		}
	}
	subscriptionByID, subscriptionByName := currentSubscriptionIndexes(nativeConfig)
	for _, key := range []string{resourceID, mihomoNodeResourcePrefix(nodeID)} {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if subscriptionByID[key].Name != "" || subscriptionByName[key].Name != "" {
			return dynamicLeaseProfileLineGroupName(profileID)
		}
	}
	return ""
}

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
