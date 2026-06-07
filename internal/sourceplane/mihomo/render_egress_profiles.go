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
	for _, profile := range opts.EgressProfiles {
		if !profile.Enabled {
			continue
		}
		id := safeID(profile.ID)
		if id == "" {
			continue
		}
		line, err := renderEgressProfileLine(opts, id, profile.Line)
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
		exit, err := renderEgressProfileExit(opts, id, profileGroupNameFor(opts, profile), profile.Exit, line.target, opts.BasePool)
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
	target       string
}

func renderEgressProfileLine(opts renderOptions, profileID string, line sourceplane.EgressProfileLine) (renderedProfileLayer, error) {
	switch strings.TrimSpace(line.Kind) {
	case "", "direct":
		return renderedProfileLayer{}, nil
	case "mihomo_node":
		return renderMihomoNativeProfileTarget(opts, profileID, line)
	default:
		return renderedProfileLayer{}, fmt.Errorf("unsupported line kind %q", line.Kind)
	}
}

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

func renderDynamicProfileExit(profileID string, groupName string, exit sourceplane.EgressProfileExit, dialerProxy string, nodes []provider.Node) (renderedProfileLayer, error) {
	group := profileLayerGroup(groupName, exit, "select")
	nodes = dynamicProfileNodes(nodes, profileID, exit.ProviderID)
	if len(nodes) == 0 {
		group.Proxies = []string{"REJECT"}
		return renderedProfileLayer{group: group}, nil
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

func dynamicProfileNodes(nodes []provider.Node, profileID string, dynamicProviderID string) []provider.Node {
	profileID = safeID(profileID)
	dynamicProviderID = strings.TrimSpace(dynamicProviderID)
	out := make([]provider.Node, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.Labels["egress_profile_id"]) != profileID {
			continue
		}
		if dynamicProviderID == "" {
			out = append(out, node)
			continue
		}
		if dynamicProfileNodeHasProviderID(node, dynamicProviderID) {
			out = append(out, node)
		}
	}
	return out
}

func dynamicProfileNodeHasProviderID(node provider.Node, dynamicProviderID string) bool {
	dynamicProviderID = strings.TrimSpace(dynamicProviderID)
	if strings.TrimSpace(node.Labels["dynamic_provider_id"]) == dynamicProviderID || strings.TrimSpace(node.ProviderID) == dynamicProviderID {
		return true
	}
	for _, value := range strings.Split(node.Labels["dynamic_provider_ids"], ",") {
		if strings.TrimSpace(value) == dynamicProviderID {
			return true
		}
	}
	return false
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

func profileInternalGroupName(profileID string) string {
	id := safeID(profileID)
	if id == "" {
		id = "profile"
	}
	return "bvf-profile-" + id
}

func profileLineGroupName(profileID string) string {
	return profileInternalGroupName(profileID) + "-line"
}

func profileGroupNames(profiles []sourceplane.EgressProfile) map[string]string {
	out := map[string]string{}
	used := map[string]int{}
	for _, profile := range profiles {
		if !profile.Enabled {
			continue
		}
		key := profileIDKey(profile.ID)
		if key == "" {
			continue
		}
		base := mihomoRuleTargetName(firstNonEmpty(profile.DisplayName, profile.ID))
		if base == "" {
			base = profileInternalGroupName(profile.ID)
		}
		name := base
		if count := used[base]; count > 0 {
			name = fmt.Sprintf("%s %d", base, count+1)
		}
		used[base]++
		out[key] = name
	}
	return out
}

func profileGroupNameFor(opts renderOptions, profile sourceplane.EgressProfile) string {
	if name := opts.ProfileGroups[profileIDKey(profile.ID)]; name != "" {
		return name
	}
	return profileInternalGroupName(profile.ID)
}

func profileIDKey(value string) string {
	return safeID(value)
}

func mihomoRuleTargetName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var out strings.Builder
	lastSpace := false
	for _, r := range value {
		if r == ',' || r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		if r == ' ' {
			if lastSpace {
				continue
			}
			lastSpace = true
			out.WriteRune(r)
			continue
		}
		lastSpace = false
		out.WriteRune(r)
	}
	return strings.TrimSpace(out.String())
}

func exactNodeFilter(name string) string {
	return "^" + regexp.QuoteMeta(strings.TrimSpace(name)) + "$"
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

func profileDynamicProxyName(profileID string, index int, node provider.Node) string {
	return safeID(fmt.Sprintf("%s-dynamic-%d-%s", profileInternalGroupName(profileID), index, providerNodeName("provider-pool", node, index)))
}

func profileGroupStrategy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "url-test", "select":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "select"
	}
}
