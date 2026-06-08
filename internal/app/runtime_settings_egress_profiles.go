package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
	"google.golang.org/protobuf/types/known/durationpb"
)

func egressProfileFromProto(in *proxyruntimev1.EgressProfileSettings) *proxyruntimev1.EgressProfileSettings {
	if in == nil {
		return &proxyruntimev1.EgressProfileSettings{}
	}
	out := &proxyruntimev1.EgressProfileSettings{
		ProfileId:   runtimeSafeID(in.GetProfileId()),
		DisplayName: strings.TrimSpace(in.GetDisplayName()),
		Enabled:     in.GetEnabled(),
		Line:        egressProfileLineFromProto(in.GetLine()),
		Exit:        egressProfileExitFromProto(in.GetExit()),
	}
	normalizeEgressProfile(out)
	return out
}

func egressProfileLineFromProto(in *proxyruntimev1.EgressProfileLineSettings) *proxyruntimev1.EgressProfileLineSettings {
	if in == nil {
		in = &proxyruntimev1.EgressProfileLineSettings{}
	}
	out := &proxyruntimev1.EgressProfileLineSettings{
		Kind:           in.GetKind(),
		MihomoNode:     egressProfileMihomoNodeRefFromProto(in.GetMihomoNode()),
		HealthCheckUrl: strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval: cloneDuration(in.GetHealthInterval()),
		HealthTimeout:  cloneDuration(in.GetHealthTimeout()),
		ExpectedStatus: in.GetExpectedStatus(),
	}
	if out.GetKind() == proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_UNSPECIFIED {
		out.Kind = proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_DIRECT
	}
	return out
}

func egressProfileExitFromProto(in *proxyruntimev1.EgressProfileExitSettings) *proxyruntimev1.EgressProfileExitSettings {
	if in == nil {
		in = &proxyruntimev1.EgressProfileExitSettings{}
	}
	out := &proxyruntimev1.EgressProfileExitSettings{
		Kind:              in.GetKind(),
		MihomoNode:        egressProfileMihomoNodeRefFromProto(in.GetMihomoNode()),
		HealthCheckUrl:    strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval:    cloneDuration(in.GetHealthInterval()),
		HealthTimeout:     cloneDuration(in.GetHealthTimeout()),
		ExpectedStatus:    in.GetExpectedStatus(),
		DynamicProviderId: strings.TrimSpace(in.GetDynamicProviderId()),
		DynamicIpPolicy:   egressProfileDynamicIPPolicyFromProto(in.GetDynamicIpPolicy()),
	}
	return out
}

func egressProfileDynamicIPPolicyFromProto(in *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	if in == nil {
		return nil
	}
	return normalizeDynamicIPSessionPolicy(in)
}

func egressProfileMihomoNodeRefFromProto(in *proxyruntimev1.EgressProfileMihomoNodeRef) *proxyruntimev1.EgressProfileMihomoNodeRef {
	if in == nil {
		return &proxyruntimev1.EgressProfileMihomoNodeRef{}
	}
	return &proxyruntimev1.EgressProfileMihomoNodeRef{
		ResourceId: strings.TrimSpace(in.GetResourceId()),
		NodeId:     strings.TrimSpace(in.GetNodeId()),
	}
}

func normalizeEgressProfile(profile *proxyruntimev1.EgressProfileSettings) {
	if profile == nil {
		return
	}
	profile.ProfileId = runtimeSafeID(profile.GetProfileId())
	profile.DisplayName = strings.TrimSpace(profile.GetDisplayName())
	profile.Line = egressProfileLineFromProto(profile.GetLine())
	profile.Exit = egressProfileExitFromProto(profile.GetExit())
}

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

type mihomoNativeResourceReplacement struct {
	ResourceID string
	FixedProxy bool
}

func (s *runtimeSettingsStore) replaceMihomoResourceRefs(ctx context.Context, replacements map[string]mihomoNativeResourceReplacement) error {
	if len(replacements) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return err
	}
	changed := false
	for _, profile := range settings.GetEgressProfiles() {
		if replaceMihomoNodeRef(profile.GetLine().GetMihomoNode(), replacements) {
			changed = true
		}
		if replaceMihomoNodeRef(profile.GetExit().GetMihomoNode(), replacements) {
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return s.saveLocked(ctx, settings)
}

func replaceMihomoNodeRef(ref *proxyruntimev1.EgressProfileMihomoNodeRef, replacements map[string]mihomoNativeResourceReplacement) bool {
	if ref == nil {
		return false
	}
	current := strings.TrimSpace(ref.GetResourceId())
	replacement, exists := replacements[current]
	if !exists {
		return false
	}
	nextResourceID := strings.TrimSpace(replacement.ResourceID)
	if nextResourceID == "" {
		return false
	}
	oldNodeID := strings.TrimSpace(ref.GetNodeId())
	ref.ResourceId = nextResourceID
	if replacement.FixedProxy {
		ref.NodeId = nextResourceID
		return current != nextResourceID || oldNodeID != nextResourceID
	}
	prefix := current + "/"
	if strings.HasPrefix(oldNodeID, prefix) {
		ref.NodeId = nextResourceID + "/" + strings.TrimPrefix(oldNodeID, prefix)
	}
	return current != nextResourceID
}

func (s *runtimeSettingsStore) enabledMihomoResourceIDs(_ context.Context) (map[string]struct{}, error) {
	return nil, nil
}

func enabledDynamicProviderIDs(settings *runtimeSettingsFile) map[string]struct{} {
	out := map[string]struct{}{}
	for _, provider := range normalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		if id := dynamicIPProviderID(provider); id != "" {
			out[id] = struct{}{}
		}
	}
	return out
}

func enabledDynamicProviderEndpointIDs(settings *runtimeSettingsFile) map[string]map[string]struct{} {
	out := map[string]map[string]struct{}{}
	for _, provider := range normalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		dynamicProviderID := dynamicIPProviderID(provider)
		if dynamicProviderID == "" {
			continue
		}
		if out[dynamicProviderID] == nil {
			out[dynamicProviderID] = map[string]struct{}{}
		}
		for _, endpoint := range provider.GetEndpoints() {
			endpointID := endpointIDFromURL(endpoint.GetEndpointUrl())
			if endpointID != "" {
				out[dynamicProviderID][endpointID] = struct{}{}
			}
		}
	}
	return out
}

func dynamicIPPolicyEndpointID(policy *proxyruntimev1.ProxySessionPolicy) string {
	return strings.TrimSpace(policy.GetLabels()["dynamic_ip_endpoint_id"])
}

func enabledEgressProfileIDsFromProfiles(profiles []*proxyruntimev1.EgressProfileSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, profile := range profiles {
		if id := runtimeSafeID(profile.GetProfileId()); id != "" && profile.GetEnabled() {
			out[id] = struct{}{}
		}
	}
	return out
}

func sourcePlaneEgressProfiles(settings *runtimeSettingsFile) []sourceplane.EgressProfile {
	settings = normalizeRuntimeSettings(settings)
	out := make([]sourceplane.EgressProfile, 0, len(settings.GetEgressProfiles()))
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		out = append(out, sourceplane.EgressProfile{
			ID:          profile.GetProfileId(),
			DisplayName: profile.GetDisplayName(),
			Enabled:     profile.GetEnabled(),
			Line:        sourcePlaneEgressProfileLine(profile.GetLine()),
			Exit:        sourcePlaneEgressProfileExit(profile.GetExit()),
		})
	}
	return out
}

func sourcePlaneEgressProfileLine(line *proxyruntimev1.EgressProfileLineSettings) sourceplane.EgressProfileLine {
	return sourceplane.EgressProfileLine{
		Kind:           egressProfileLineKind(line.GetKind()),
		ResourceID:     strings.TrimSpace(line.GetMihomoNode().GetResourceId()),
		NodeID:         strings.TrimSpace(line.GetMihomoNode().GetNodeId()),
		HealthCheckURL: strings.TrimSpace(line.GetHealthCheckUrl()),
		HealthInterval: protoDuration(line.GetHealthInterval(), 300*time.Second),
		HealthTimeout:  protoDuration(line.GetHealthTimeout(), 5*time.Second),
		ExpectedStatus: defaultExpectedStatus(line.GetExpectedStatus()),
	}
}

func sourcePlaneEgressProfileExit(exit *proxyruntimev1.EgressProfileExitSettings) sourceplane.EgressProfileExit {
	return sourceplane.EgressProfileExit{
		Kind:           egressProfileExitKind(exit.GetKind()),
		ProviderID:     strings.TrimSpace(exit.GetDynamicProviderId()),
		ResourceID:     strings.TrimSpace(exit.GetMihomoNode().GetResourceId()),
		NodeID:         strings.TrimSpace(exit.GetMihomoNode().GetNodeId()),
		HealthCheckURL: strings.TrimSpace(exit.GetHealthCheckUrl()),
		HealthInterval: protoDuration(exit.GetHealthInterval(), 300*time.Second),
		HealthTimeout:  protoDuration(exit.GetHealthTimeout(), 5*time.Second),
		ExpectedStatus: defaultExpectedStatus(exit.GetExpectedStatus()),
	}
}

func egressProfileLineKind(kind proxyruntimev1.EgressProfileLineKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE:
		return "mihomo_node"
	default:
		return "direct"
	}
}

func egressProfileExitKind(kind proxyruntimev1.EgressProfileExitKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DIRECT:
		return "direct"
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
		return "static_ip"
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
		return "dynamic_ip"
	default:
		return ""
	}
}

func cloneDuration(value *durationpb.Duration) *durationpb.Duration {
	if value == nil {
		return nil
	}
	return durationpb.New(value.AsDuration())
}
