package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

func renderMihomoNativeProfileTarget(opts renderOptions, profileID string, layer sourceplane.EgressProfileLayer) (renderedProfileLayer, error) {
	target, err := resolveMihomoNativeLayerTarget(opts.NativeConfig, layer)
	if err != nil {
		return renderedProfileLayer{}, err
	}
	if mihomoProxyAvailable(opts.AvailableProxies, target.NodeName) {
		return renderedProfileLayer{target: target.NodeName}, nil
	}
	if mihomoProviderAvailable(opts.AvailableProviders, target.ResourceName) {
		group := profileLayerGroup(profileLineGroupName(profileID), layer, "select")
		group.Use = []string{target.ResourceName}
		group.Filter = exactNodeFilter(target.NodeName)
		return renderedProfileLayer{group: group, target: group.Name}, nil
	}
	return renderedProfileLayer{target: "REJECT"}, nil
}

func renderMihomoNativeProfileLayer(opts renderOptions, groupName string, layer sourceplane.EgressProfileLayer, dialerProxy string) (renderedProfileLayer, error) {
	target, err := resolveMihomoNativeLayerTarget(opts.NativeConfig, layer)
	if err != nil {
		return renderedProfileLayer{}, err
	}
	if strings.TrimSpace(dialerProxy) != "" {
		return renderedProfileLayer{}, fmt.Errorf("mihomo-native node %q cannot be cloned with dialer-proxy by proxy-gateway", target.NodeName)
	}
	group := profileLayerGroup(groupName, layer, "select")
	if mihomoProxyAvailable(opts.AvailableProxies, target.NodeName) {
		group.Proxies = []string{target.NodeName}
		return renderedProfileLayer{group: group}, nil
	}
	if mihomoProviderAvailable(opts.AvailableProviders, target.ResourceName) {
		group.Use = []string{target.ResourceName}
		group.Filter = exactNodeFilter(target.NodeName)
		return renderedProfileLayer{group: group}, nil
	}
	group.Proxies = []string{"REJECT"}
	return renderedProfileLayer{group: group}, nil
}
