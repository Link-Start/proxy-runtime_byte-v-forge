package app

import (
	"context"
	"fmt"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"
)

func (r *Runtime) dynamicLeaseDialerProxy(ctx context.Context, settings *runtimeSettingsFile, profileID string) (string, map[string]string, error) {
	settings = kernel.NormalizeRuntimeSettings(settings)
	profiles := dynamicLeaseLineProfiles(settings, profileID)
	if len(profiles) == 0 {
		return "", nil, nil
	}
	nativeSettings, err := r.settings.LoadMihomoNative(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("load mihomo native settings for lease line: %w", err)
	}
	nativeConfig, err := mihomonative.ConfigFromSettings(nativeSettings)
	if err != nil {
		return "", nil, fmt.Errorf("render mihomo native settings for lease line: %w", err)
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

func dynamicLeaseProfileDialerProxy(profile *proxygatewayv1.EgressProfileSettings, nativeConfig mihomonative.ConfigFile) (string, map[string]string, error) {
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

func dynamicLeaseLineDialerProxy(profileID string, nativeConfig mihomonative.ConfigFile, resourceID string, nodeID string) string {
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

func mihomoNodeResourcePrefix(nodeID string) string {
	resource, _, ok := strings.Cut(strings.TrimSpace(nodeID), "/")
	if !ok {
		return ""
	}
	return strings.TrimSpace(resource)
}

func dynamicLeaseProfileLineGroupName(profileID string) string {
	id := appcore.RuntimeSafeID(profileID)
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

func dynamicLeaseNativeDialerProxy(profileID string, nativeConfig mihomonative.ConfigFile, resourceID string, nodeID string) string {
	fixedByID, fixedByName := mihomonative.CurrentFixedProxyIndexes(nativeConfig)
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
	subscriptionByID, subscriptionByName := mihomonative.CurrentSubscriptionIndexes(nativeConfig)
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

func dynamicLeaseLineProfiles(settings *runtimeSettingsFile, profileID string) []*proxygatewayv1.EgressProfileSettings {
	profileID = appcore.RuntimeSafeID(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		if appcore.RuntimeSafeID(profile.GetProfileId()) != profileID {
			continue
		}
		if profile.GetExit().GetKind() != proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		line := profile.GetLine()
		if line.GetKind() != proxygatewayv1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE {
			continue
		}
		return []*proxygatewayv1.EgressProfileSettings{profile}
	}
	return nil
}
