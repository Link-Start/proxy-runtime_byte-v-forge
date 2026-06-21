package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

type mihomoNativeLayerTarget struct {
	ResourceName string
	NodeName     string
}

func resolveMihomoNativeLayerTarget(config mihomoNativeConfig, layer sourceplane.EgressProfileLayer) (mihomoNativeLayerTarget, error) {
	resourceID := strings.TrimSpace(layer.ResourceID)
	if resourceID == "" {
		return mihomoNativeLayerTarget{}, fmt.Errorf("resource_id is required")
	}
	resolved := resolveMihomoNativeResource(config, resourceID, layer.NodeID)
	resourceName := firstNonEmpty(resolved.ResourceName, resourceID)
	nodeName := firstNonEmpty(resolved.NodeName, mihomoNativeNodeName(resourceName, layer.NodeID), mihomoNativeNodeName(resourceID, layer.NodeID))
	if nodeName == "" {
		return mihomoNativeLayerTarget{}, fmt.Errorf("node_id is required")
	}
	return mihomoNativeLayerTarget{ResourceName: resourceName, NodeName: nodeName}, nil
}
