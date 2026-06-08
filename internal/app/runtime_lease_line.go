package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func (r *Runtime) dynamicLeaseDialerProxy(settings *runtimeSettingsFile, profileID string) (string, map[string]string, error) {
	settings = normalizeRuntimeSettings(settings)
	profiles := dynamicLeaseLineProfiles(settings, profileID)
	if len(profiles) == 0 {
		return "", nil, nil
	}
	nativeConfig, err := loadMihomoNativeConfig(r)
	if err != nil {
		return "", nil, fmt.Errorf("load mihomo native config for lease line: %w", err)
	}
	for _, profile := range profiles {
		dialer, labels, err := dynamicLeaseProfileDialerProxy(profile, nativeConfig)
		if err != nil {
			return "", nil, err
		}
		if dialer != "" {
			return dialer, labels, nil
		}
	}
	return "", nil, nil
}

func dynamicLeaseLineProfiles(settings *runtimeSettingsFile, profileID string) []*proxyruntimev1.EgressProfileSettings {
	profileID = runtimeSafeID(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		if runtimeSafeID(profile.GetProfileId()) != profileID {
			continue
		}
		if profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		line := profile.GetLine()
		if line.GetKind() != proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE {
			continue
		}
		return []*proxyruntimev1.EgressProfileSettings{profile}
	}
	return nil
}

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

func applyDynamicLeaseLineLabels(nodes []provider.Node, labels map[string]string) []provider.Node {
	if len(labels) == 0 {
		return nodes
	}
	for index := range nodes {
		nodes[index].Labels = cloneLabels(nodes[index].Labels)
		if nodes[index].Labels == nil {
			nodes[index].Labels = map[string]string{}
		}
		for key, value := range labels {
			nodes[index].Labels[key] = value
		}
	}
	return nodes
}
