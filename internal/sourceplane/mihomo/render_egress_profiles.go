package mihomo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func renderEgressProfiles(opts renderOptions) ([]map[string]any, map[string]mihomoProvider, []mihomoGroup, error) {
	proxies := []map[string]any{}
	providers := map[string]mihomoProvider{}
	groups := []mihomoGroup{}
	subscriptions := subscriptionProviderMap(opts.Providers)
	fixed := fixedProxyMap(opts.FixedProxies)
	for _, profile := range opts.EgressProfiles {
		if !profile.Enabled {
			continue
		}
		id := safeID(profile.ID)
		if id == "" {
			continue
		}
		line, err := renderEgressProfileLine(opts, id, profile.Line, subscriptions, fixed)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("profile %q line: %w", id, err)
		}
		if line.proxy != nil {
			proxies = append(proxies, line.proxy)
		}
		if line.providerName != "" {
			providers[line.providerName] = line.provider
		}
		if line.group.Name != "" {
			groups = append(groups, line.group)
		}
		exit, err := renderEgressProfileExit(opts, id, profile.Exit, line.group.Name, subscriptions, fixed)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("profile %q exit: %w", id, err)
		}
		if exit.proxy != nil {
			proxies = append(proxies, exit.proxy)
		}
		proxies = append(proxies, exit.proxies...)
		if exit.providerName != "" {
			providers[exit.providerName] = exit.provider
		}
		groups = append(groups, exit.group)
	}
	return proxies, providers, groups, nil
}

type renderedProfileLayer struct {
	proxy        map[string]any
	proxies      []map[string]any
	providerName string
	provider     mihomoProvider
	group        mihomoGroup
}

func renderEgressProfileLine(opts renderOptions, profileID string, line sourceplane.EgressProfileLine, subscriptions map[string]sourceplane.SubscriptionProvider, fixed map[string]sourceplane.FixedProxy) (renderedProfileLayer, error) {
	switch strings.TrimSpace(line.Kind) {
	case "", "direct":
		return renderedProfileLayer{}, nil
	case "source":
		return renderSourceProfileLayer(opts, profileID, "line", profileLineGroupName(profileID), line, "", subscriptions, fixed)
	default:
		return renderedProfileLayer{}, fmt.Errorf("unsupported line kind %q", line.Kind)
	}
}

func renderEgressProfileExit(opts renderOptions, profileID string, exit sourceplane.EgressProfileExit, dialerProxy string, subscriptions map[string]sourceplane.SubscriptionProvider, fixed map[string]sourceplane.FixedProxy) (renderedProfileLayer, error) {
	switch strings.TrimSpace(exit.Kind) {
	case "direct":
		return renderDirectProfileExit(profileID, exit, dialerProxy)
	case "static_ip":
		return renderStaticProfileExit(opts, profileID, exit, dialerProxy, subscriptions, fixed)
	case "dynamic_ip":
		return renderDynamicProfileExit(profileID, exit, dialerProxy, opts.BasePool)
	default:
		return renderedProfileLayer{}, fmt.Errorf("unsupported exit kind %q", exit.Kind)
	}
}

func renderSourceProfileLayer(opts renderOptions, profileID string, layerID string, groupName string, layer sourceplane.EgressProfileLayer, dialerProxy string, subscriptions map[string]sourceplane.SubscriptionProvider, fixed map[string]sourceplane.FixedProxy) (renderedProfileLayer, error) {
	sourceID := safeID(layer.SourceID)
	if _, exists := fixed[sourceID]; exists {
		return renderFixedProfileLayer(profileID, layerID, groupName, layer, dialerProxy, fixed)
	}
	if _, exists := subscriptions[sourceID]; exists {
		return renderSubscriptionProfileLayer(opts, profileID, layerID, groupName, layer, dialerProxy, subscriptions)
	}
	return renderedProfileLayer{}, fmt.Errorf("source %q is not enabled", sourceID)
}

func renderDirectProfileExit(profileID string, exit sourceplane.EgressProfileExit, dialerProxy string) (renderedProfileLayer, error) {
	group := profileLayerGroup(profileGroupName(profileID), exit, "select")
	if strings.TrimSpace(dialerProxy) == "" {
		group.Proxies = []string{"DIRECT"}
		return renderedProfileLayer{group: group}, nil
	}
	group.Proxies = []string{strings.TrimSpace(dialerProxy)}
	return renderedProfileLayer{group: group}, nil
}

func renderStaticProfileExit(opts renderOptions, profileID string, exit sourceplane.EgressProfileExit, dialerProxy string, subscriptions map[string]sourceplane.SubscriptionProvider, fixed map[string]sourceplane.FixedProxy) (renderedProfileLayer, error) {
	sourceID := safeID(exit.SourceID)
	if _, exists := fixed[sourceID]; exists {
		return renderFixedProfileLayer(profileID, "exit", profileGroupName(profileID), exit, dialerProxy, fixed)
	}
	if _, exists := subscriptions[sourceID]; exists {
		return renderSubscriptionProfileLayer(opts, profileID, "exit", profileGroupName(profileID), exit, dialerProxy, subscriptions)
	}
	return renderedProfileLayer{}, fmt.Errorf("source %q is not enabled", sourceID)
}

func renderFixedProfileLayer(profileID string, layerID string, groupName string, layer sourceplane.EgressProfileLayer, dialerProxy string, fixed map[string]sourceplane.FixedProxy) (renderedProfileLayer, error) {
	sourceID := safeID(layer.SourceID)
	if sourceID == "" {
		return renderedProfileLayer{}, fmt.Errorf("source_id is required")
	}
	if strings.TrimSpace(layer.NodeID) == "" {
		return renderedProfileLayer{}, fmt.Errorf("node_id is required")
	}
	group := profileLayerGroup(groupName, layer, "select")
	if item, exists := fixed[sourceID]; exists {
		proxyName := sourceID
		var proxy map[string]any
		if strings.TrimSpace(dialerProxy) != "" {
			proxyName = profileProxyName(profileID, layerID, sourceID)
			item.ID = proxyName
			var err error
			proxy, err = renderFixedProxy(item)
			if err != nil {
				return renderedProfileLayer{}, err
			}
			proxy["dialer-proxy"] = strings.TrimSpace(dialerProxy)
		}
		group.Proxies = []string{fixedProfileProxyTarget(layer, proxyName, dialerProxy)}
		return renderedProfileLayer{proxy: proxy, group: group}, nil
	}
	return renderedProfileLayer{}, fmt.Errorf("fixed proxy source %q is not enabled", sourceID)
}

func renderSubscriptionProfileLayer(opts renderOptions, profileID string, layerID string, groupName string, layer sourceplane.EgressProfileLayer, dialerProxy string, subscriptions map[string]sourceplane.SubscriptionProvider) (renderedProfileLayer, error) {
	sourceID := safeID(layer.SourceID)
	if sourceID == "" {
		return renderedProfileLayer{}, fmt.Errorf("source_id is required")
	}
	if strings.TrimSpace(layer.NodeID) == "" {
		return renderedProfileLayer{}, fmt.Errorf("node_id is required")
	}
	group := profileLayerGroup(groupName, layer, "select")
	if item, exists := subscriptions[sourceID]; exists {
		providerName := sourceID
		var provider mihomoProvider
		if strings.TrimSpace(dialerProxy) != "" {
			providerName = profileProviderName(profileID, layerID, sourceID)
			provider = renderSubscriptionProvider(opts, item, providerName, dialerProxy)
		}
		group.Use = []string{providerName}
		nodeID := strings.TrimSpace(layer.NodeID)
		nodeName := providerProxyNameFromNodeID(providerFileCandidates(opts.ConfigDir, item, sourceID), nodeID)
		if nodeName == "" {
			return renderedProfileLayer{}, fmt.Errorf("subscription source %q node %q is not available", sourceID, nodeID)
		}
		if strings.TrimSpace(dialerProxy) != "" {
			nodeName = safeID(providerName) + "-" + nodeName
		}
		group.Filter = exactNodeFilter(nodeName)
		return renderedProfileLayer{providerName: providerNameForReturn(providerName, sourceID, dialerProxy), provider: provider, group: group}, nil
	}
	return renderedProfileLayer{}, fmt.Errorf("subscription source %q is not enabled", sourceID)
}

func renderDynamicProfileExit(profileID string, exit sourceplane.EgressProfileExit, dialerProxy string, nodes []provider.Node) (renderedProfileLayer, error) {
	group := profileLayerGroup(profileGroupName(profileID), exit, "url-test")
	nodes = dynamicProfileNodes(nodes, exit.ProviderID)
	if len(nodes) == 0 {
		if strings.TrimSpace(exit.ProviderID) == "" {
			return renderedProfileLayer{}, fmt.Errorf("dynamic provider pool is empty")
		}
		return renderedProfileLayer{}, fmt.Errorf("dynamic provider pool %q is empty", exit.ProviderID)
	}
	if strings.TrimSpace(dialerProxy) == "" {
		group.Proxies = make([]string, 0, len(nodes))
		for index, node := range nodes {
			group.Proxies = append(group.Proxies, providerNodeName("provider-pool", node, index))
		}
		return renderedProfileLayer{group: group}, nil
	}
	out := renderedProfileLayer{group: group, proxies: make([]map[string]any, 0, len(nodes))}
	for index, node := range nodes {
		name := profileDynamicProxyName(profileID, index, node)
		proxy, err := renderProxyURL(name, node.URL)
		if err != nil {
			return renderedProfileLayer{}, err
		}
		proxy["dialer-proxy"] = strings.TrimSpace(dialerProxy)
		out.proxies = append(out.proxies, proxy)
		out.group.Proxies = append(out.group.Proxies, name)
	}
	return out, nil
}

func dynamicProfileNodes(nodes []provider.Node, providerID string) []provider.Node {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return nodes
	}
	out := make([]provider.Node, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.ProviderID) == providerID {
			out = append(out, node)
		}
	}
	return out
}

func profileLayerGroup(name string, layer sourceplane.EgressProfileLayer, groupType string) mihomoGroup {
	return mihomoGroup{
		Name:           name,
		Type:           profileGroupStrategy(groupType),
		URL:            firstNonEmpty(layer.HealthCheckURL, "https://www.gstatic.com/generate_204"),
		Interval:       seconds(layer.HealthInterval, 300),
		Timeout:        milliseconds(layer.HealthTimeout, 5000),
		Lazy:           true,
		ExpectedStatus: defaultExpectedStatus(layer.ExpectedStatus),
		Hidden:         true,
	}
}

func providerNameForReturn(providerName string, sourceID string, dialerProxy string) string {
	if strings.TrimSpace(dialerProxy) == "" || providerName == sourceID {
		return ""
	}
	return providerName
}

func fixedProfileProxyTarget(layer sourceplane.EgressProfileLayer, proxyName string, dialerProxy string) string {
	if strings.TrimSpace(dialerProxy) != "" {
		return safeID(proxyName)
	}
	return safeID(layer.NodeID)
}

func exactNodeFilter(name string) string {
	return "^" + regexp.QuoteMeta(strings.TrimSpace(name)) + "$"
}

func subscriptionProviderMap(items []sourceplane.SubscriptionProvider) map[string]sourceplane.SubscriptionProvider {
	out := map[string]sourceplane.SubscriptionProvider{}
	for _, item := range items {
		if id := safeID(item.ID); id != "" {
			out[id] = item
		}
	}
	return out
}

func fixedProxyMap(items []sourceplane.FixedProxy) map[string]sourceplane.FixedProxy {
	out := map[string]sourceplane.FixedProxy{}
	for _, item := range items {
		if id := safeID(item.ID); id != "" {
			out[id] = item
		}
	}
	return out
}

func profileGroupName(profileID string) string {
	id := safeID(profileID)
	if id == "" {
		id = "profile"
	}
	return "bvf-profile-" + id
}

func profileLineGroupName(profileID string) string {
	return profileGroupName(profileID) + "-line"
}

func profileProxyName(profileID string, layerID string, sourceID string) string {
	return safeID(fmt.Sprintf("%s-%s-proxy-%s", profileGroupName(profileID), layerID, safeID(sourceID)))
}

func profileProviderName(profileID string, layerID string, sourceID string) string {
	return safeID(fmt.Sprintf("%s-%s-provider-%s", profileGroupName(profileID), layerID, safeID(sourceID)))
}

func profileDynamicProxyName(profileID string, index int, node provider.Node) string {
	return safeID(fmt.Sprintf("%s-dynamic-%d-%s", profileGroupName(profileID), index, providerNodeName("provider-pool", node, index)))
}

func profileGroupStrategy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "url-test", "select":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "select"
	}
}
