package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func renderMihomoNativeProfileTarget(opts renderOptions, profileID string, layer sourceplane.EgressProfileLayer) (renderedProfileLayer, error) {
	resourceID := strings.TrimSpace(layer.ResourceID)
	if resourceID == "" {
		return renderedProfileLayer{}, fmt.Errorf("resource_id is required")
	}
	resolved := resolveMihomoNativeResource(opts.NativeConfig, resourceID, layer.NodeID)
	resourceName := firstNonEmpty(resolved.ResourceName, resourceID)
	nodeName := firstNonEmpty(resolved.NodeName, mihomoNativeNodeName(resourceName, layer.NodeID), mihomoNativeNodeName(resourceID, layer.NodeID))
	if nodeName == "" {
		return renderedProfileLayer{}, fmt.Errorf("node_id is required")
	}
	if mihomoProxyAvailable(opts.AvailableProxies, nodeName) {
		return renderedProfileLayer{target: nodeName}, nil
	}
	if mihomoProviderAvailable(opts.AvailableProviders, resourceName) {
		group := profileLayerGroup(profileLineGroupName(profileID), layer, "select")
		group.Use = []string{resourceName}
		group.Filter = exactNodeFilter(nodeName)
		return renderedProfileLayer{group: group, target: group.Name}, nil
	}
	return renderedProfileLayer{target: "REJECT"}, nil
}

func renderMihomoNativeProfileLayer(opts renderOptions, groupName string, layer sourceplane.EgressProfileLayer, dialerProxy string) (renderedProfileLayer, error) {
	resourceID := strings.TrimSpace(layer.ResourceID)
	if resourceID == "" {
		return renderedProfileLayer{}, fmt.Errorf("resource_id is required")
	}
	resolved := resolveMihomoNativeResource(opts.NativeConfig, resourceID, layer.NodeID)
	resourceName := firstNonEmpty(resolved.ResourceName, resourceID)
	nodeName := firstNonEmpty(resolved.NodeName, mihomoNativeNodeName(resourceName, layer.NodeID), mihomoNativeNodeName(resourceID, layer.NodeID))
	if nodeName == "" {
		return renderedProfileLayer{}, fmt.Errorf("node_id is required")
	}
	if strings.TrimSpace(dialerProxy) != "" {
		return renderedProfileLayer{}, fmt.Errorf("mihomo-native node %q cannot be cloned with dialer-proxy by proxy-runtime", nodeName)
	}
	group := profileLayerGroup(groupName, layer, "select")
	if mihomoProxyAvailable(opts.AvailableProxies, nodeName) {
		group.Proxies = []string{nodeName}
		return renderedProfileLayer{group: group}, nil
	}
	if mihomoProviderAvailable(opts.AvailableProviders, resourceName) {
		group.Use = []string{resourceName}
		group.Filter = exactNodeFilter(nodeName)
		return renderedProfileLayer{group: group}, nil
	}
	group.Proxies = []string{"REJECT"}
	return renderedProfileLayer{group: group}, nil
}
