package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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

func (s *runtimeSettingsStore) updateEgressProfiles(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
	if err != nil {
		return nil, err
	}
	nextProfiles, err := egressProfilesFromRequest(profiles, nativeResourceIDs, enabledDynamicProviderEndpointIDs(settings))
	if err != nil {
		return nil, err
	}
	nextRules, err := ingressRulesFromRequest(settings.GetIngressRules(), nextProfiles)
	if err != nil {
		return nil, err
	}
	applyInUserSessionLabels(nextProfiles, nextRules)
	settings.EgressProfiles = nextProfiles
	settings.IngressRules = nextRules
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) updateIngressRules(ctx context.Context, rules []*proxyruntimev1.ProxyIngressRuleSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	nextRules, err := ingressRulesFromRequest(rules, settings.GetEgressProfiles())
	if err != nil {
		return nil, err
	}
	applyInUserSessionLabels(settings.EgressProfiles, nextRules)
	settings.IngressRules = nextRules
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

type mihomoNativeResourceReplacement struct {
	ResourceID string
	FixedProxy bool
}

func (s *runtimeSettingsStore) replaceMihomoResourceRefs(ctx context.Context, replacements map[string]mihomoNativeResourceReplacement) (bool, error) {
	if len(replacements) == 0 {
		return false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return false, err
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
		return false, nil
	}
	if err := s.saveLocked(ctx, settings); err != nil {
		return false, err
	}
	return true, nil
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

func (s *runtimeSettingsStore) enabledMihomoResourceIDs(ctx context.Context) (map[string]struct{}, error) {
	view, err := s.loadMihomoNativeLocked(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	for _, proxy := range view.GetFixedProxies() {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(proxy), nil)
		addEnabledMihomoResourceID(out, normalized.ID)
		addEnabledMihomoResourceID(out, normalized.Name)
	}
	for _, subscription := range view.GetSubscriptions() {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(subscription), nil)
		addEnabledMihomoResourceID(out, normalized.ID)
		addEnabledMihomoResourceID(out, normalized.Name)
	}
	return out, nil
}

func addEnabledMihomoResourceID(resources map[string]struct{}, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		resources[value] = struct{}{}
	}
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

func enabledEgressProfileIDsFromProfiles(profiles []*proxyruntimev1.EgressProfileSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, profile := range profiles {
		if id := runtimeSafeID(profile.GetProfileId()); id != "" && profile.GetEnabled() {
			out[id] = struct{}{}
		}
	}
	return out
}

func cloneDuration(value *durationpb.Duration) *durationpb.Duration {
	if value == nil {
		return nil
	}
	return durationpb.New(value.AsDuration())
}
