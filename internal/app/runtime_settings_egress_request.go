package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func validateEgressProfile(profile *proxyruntimev1.EgressProfileSettings, index int, nativeResourceIDs map[string]struct{}, dynamicProviderEndpointIDs map[string]map[string]struct{}) error {
	if profile.GetProfileId() == "" {
		return fmt.Errorf("egress_profiles[%d].profile_id is required", index)
	}
	if err := validateEgressProfileLine(profile.GetLine(), index, nativeResourceIDs); err != nil {
		return err
	}
	return validateEgressProfileExit(profile.GetLine(), profile.GetExit(), index, nativeResourceIDs, dynamicProviderEndpointIDs)
}

func validateEgressProfileLine(line *proxyruntimev1.EgressProfileLineSettings, index int, nativeResourceIDs map[string]struct{}) error {
	switch line.GetKind() {
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_DIRECT:
		return nil
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE:
		return validateEgressProfileResource(line.GetMihomoNode(), fmt.Sprintf("egress_profiles[%d].line.mihomo_node", index), nativeResourceIDs, true)
	default:
		return fmt.Errorf("egress_profiles[%d].line.kind is required", index)
	}
}

func validateEgressProfileExit(line *proxyruntimev1.EgressProfileLineSettings, exit *proxyruntimev1.EgressProfileExitSettings, index int, nativeResourceIDs map[string]struct{}, dynamicProviderEndpointIDs map[string]map[string]struct{}) error {
	switch exit.GetKind() {
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DIRECT:
		return nil
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
		if err := validateEgressProfileResource(exit.GetMihomoNode(), fmt.Sprintf("egress_profiles[%d].exit.mihomo_node", index), nativeResourceIDs, true); err != nil {
			return err
		}
		if line.GetKind() == proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE {
			return fmt.Errorf("egress_profiles[%d].exit static_ip requires direct line because Mihomo-native nodes are not cloned by proxy-runtime", index)
		}
		return nil
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
		dynamicProviderID := strings.TrimSpace(exit.GetDynamicProviderId())
		endpointID := dynamicIPPolicyEndpointID(exit.GetDynamicIpPolicy())
		if dynamicProviderID == "" {
			if endpointID != "" {
				return fmt.Errorf("egress_profiles[%d].exit.dynamic_ip_policy.labels.dynamic_ip_endpoint_id requires dynamic_provider_id", index)
			}
			return nil
		}
		endpoints, exists := dynamicProviderEndpointIDs[dynamicProviderID]
		if dynamicProviderEndpointIDs != nil && !exists {
			return fmt.Errorf("egress_profiles[%d].exit.dynamic_provider_id %q is not enabled", index, dynamicProviderID)
		}
		if endpointID == "" {
			return nil
		}
		if _, exists := endpoints[endpointID]; !exists {
			return fmt.Errorf("egress_profiles[%d].exit.dynamic_ip_policy.labels.dynamic_ip_endpoint_id %q is not enabled for dynamic_provider_id %q", index, endpointID, dynamicProviderID)
		}
		return nil
	default:
		return fmt.Errorf("egress_profiles[%d].exit.kind is required", index)
	}
}

func validateEgressProfileResource(resource *proxyruntimev1.EgressProfileMihomoNodeRef, field string, nativeResourceIDs map[string]struct{}, requireNode bool) error {
	resourceID := strings.TrimSpace(resource.GetResourceId())
	if resourceID == "" {
		return fmt.Errorf("%s.resource_id is required", field)
	}
	if requireNode && strings.TrimSpace(resource.GetNodeId()) == "" {
		return fmt.Errorf("%s.node_id is required", field)
	}
	if nativeResourceIDs != nil {
		if _, exists := nativeResourceIDs[resourceID]; !exists {
			return fmt.Errorf("%s.resource_id %q is not enabled", field, resourceID)
		}
	}
	return nil
}

func cloneEgressProfile(in *proxyruntimev1.EgressProfileSettings) *proxyruntimev1.EgressProfileSettings {
	return egressProfileFromProto(in)
}

func egressProfilesFromRequest(in []*proxyruntimev1.EgressProfileSettings, nativeResourceIDs map[string]struct{}, dynamicProviderEndpointIDs map[string]map[string]struct{}) ([]*proxyruntimev1.EgressProfileSettings, error) {
	out := make([]*proxyruntimev1.EgressProfileSettings, 0, len(in))
	seen := map[string]struct{}{}
	for index, profile := range in {
		item := egressProfileFromProto(profile)
		if err := validateEgressProfile(item, index, nativeResourceIDs, dynamicProviderEndpointIDs); err != nil {
			return nil, err
		}
		if _, exists := seen[item.GetProfileId()]; exists {
			return nil, fmt.Errorf("egress_profiles[%d] duplicates profile %q", index, item.GetProfileId())
		}
		seen[item.GetProfileId()] = struct{}{}
		out = append(out, item)
	}
	return out, nil
}

func dynamicIPPolicyEndpointID(policy *proxyruntimev1.ProxySessionPolicy) string {
	return strings.TrimSpace(policy.GetLabels()["dynamic_ip_endpoint_id"])
}
