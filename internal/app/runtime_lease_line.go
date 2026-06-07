package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func dynamicLeaseDialerProxy(settings *runtimeSettingsFile, profileID string, selected *proxyruntimev1.ProxyDynamicIPEndpointCandidate) (string, map[string]string) {
	settings = normalizeRuntimeSettings(settings)
	profiles := dynamicLeaseLineProfiles(settings, profileID, selected)
	for _, profile := range profiles {
		dialer, labels := dynamicLeaseProfileDialerProxy(profile)
		if dialer != "" {
			return dialer, labels
		}
	}
	return "", nil
}

func dynamicLeaseLineProfiles(settings *runtimeSettingsFile, profileID string, selected *proxyruntimev1.ProxyDynamicIPEndpointCandidate) []*proxyruntimev1.EgressProfileSettings {
	selectedDynamicProviderID := runtimeSafeID(selected.GetDynamicProviderId())
	profileID = runtimeSafeID(profileID)
	exact := make([]*proxyruntimev1.EgressProfileSettings, 0, 1)
	fallback := make([]*proxyruntimev1.EgressProfileSettings, 0)
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		if profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		profileDynamicProviderID := runtimeSafeID(profile.GetExit().GetDynamicProviderId())
		if profileDynamicProviderID != "" && selectedDynamicProviderID != "" && profileDynamicProviderID != selectedDynamicProviderID {
			continue
		}
		line := profile.GetLine()
		if line.GetKind() != proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE {
			continue
		}
		if profileID != "" && runtimeSafeID(profile.GetProfileId()) == profileID {
			exact = append(exact, profile)
			continue
		}
		fallback = append(fallback, profile)
	}
	if len(exact) > 0 {
		return exact
	}
	return fallback
}

func dynamicLeaseProfileDialerProxy(profile *proxyruntimev1.EgressProfileSettings) (string, map[string]string) {
	node := profile.GetLine().GetMihomoNode()
	dialer := dynamicLeaseLineDialerProxy(profile.GetProfileId(), node.GetResourceId(), node.GetNodeId())
	if dialer == "" {
		return "", nil
	}
	return dialer, map[string]string{
		"line_kind":        "mihomo_node",
		"line_resource_id": strings.TrimSpace(node.GetResourceId()),
		"line_node_id":     strings.TrimSpace(node.GetNodeId()),
		"line_dialer":      dialer,
	}
}

func dynamicLeaseLineDialerProxy(profileID string, resourceID string, nodeID string) string {
	resourceID = strings.TrimSpace(resourceID)
	nodeID = strings.TrimSpace(nodeID)
	if resourceID != "" && strings.HasPrefix(nodeID, resourceID+"/") {
		return dynamicLeaseProfileLineGroupName(profileID)
	}
	return mihomoNodeDialerProxyName(resourceID, nodeID)
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
