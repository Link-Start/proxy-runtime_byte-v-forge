package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-gateway/internal/provider"
	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

func renderEgressProfileExit(opts renderOptions, profileID string, groupName string, exit sourceplane.EgressProfileExit, dialerProxy string, nodes []provider.Node) (renderedProfileLayer, error) {
	switch strings.TrimSpace(exit.Kind) {
	case "direct":
		return renderDirectProfileExit(groupName, exit, dialerProxy)
	case "static_ip":
		return renderMihomoNativeProfileLayer(opts, groupName, exit, dialerProxy)
	case "dynamic_ip":
		return renderDynamicProfileExit(profileID, groupName, exit, dialerProxy, nodes)
	default:
		return renderedProfileLayer{}, fmt.Errorf("unsupported exit kind %q", exit.Kind)
	}
}

func renderDirectProfileExit(groupName string, exit sourceplane.EgressProfileExit, dialerProxy string) (renderedProfileLayer, error) {
	group := profileLayerGroup(groupName, exit, "select")
	if strings.TrimSpace(dialerProxy) == "" {
		group.Proxies = []string{"DIRECT"}
		return renderedProfileLayer{group: group}, nil
	}
	group.Proxies = []string{strings.TrimSpace(dialerProxy)}
	return renderedProfileLayer{group: group}, nil
}
