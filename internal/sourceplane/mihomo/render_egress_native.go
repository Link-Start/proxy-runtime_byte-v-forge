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
