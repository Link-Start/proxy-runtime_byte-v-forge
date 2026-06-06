package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func dynamicLeaseDialerProxy(settings *runtimeSettingsFile, selected *proxyruntimev1.ProxyDynamicIPEndpointCandidate) (string, map[string]string) {
	settings = normalizeRuntimeSettings(settings)
	selectedDynamicProviderID := runtimeSafeID(selected.GetDynamicProviderId())
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
		node := line.GetMihomoNode()
		dialer := mihomoNodeDialerProxyName(node.GetResourceId(), node.GetNodeId())
		if dialer == "" {
			continue
		}
		return dialer, map[string]string{
			"line_kind":        "mihomo_node",
			"line_resource_id": strings.TrimSpace(node.GetResourceId()),
			"line_node_id":     strings.TrimSpace(node.GetNodeId()),
			"line_dialer":      dialer,
		}
	}
	return "", nil
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
