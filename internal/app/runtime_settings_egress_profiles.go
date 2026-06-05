package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
	"google.golang.org/protobuf/types/known/durationpb"
)

func egressProfileFromProto(in *proxyruntimev1.EgressProfileSettings) *proxyruntimev1.EgressProfileSettings {
	if in == nil {
		return &proxyruntimev1.EgressProfileSettings{}
	}
	out := &proxyruntimev1.EgressProfileSettings{
		ProfileId:   sourceSafeID(in.GetProfileId()),
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
		Source:         egressProfileSourceRefFromProto(in.GetSource()),
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
		Source:            egressProfileSourceRefFromProto(in.GetSource()),
		HealthCheckUrl:    strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval:    cloneDuration(in.GetHealthInterval()),
		HealthTimeout:     cloneDuration(in.GetHealthTimeout()),
		ExpectedStatus:    in.GetExpectedStatus(),
		DynamicProviderId: strings.TrimSpace(in.GetDynamicProviderId()),
	}
	return out
}

func egressProfileSourceRefFromProto(in *proxyruntimev1.EgressProfileSourceRef) *proxyruntimev1.EgressProfileSourceRef {
	if in == nil {
		return &proxyruntimev1.EgressProfileSourceRef{}
	}
	return &proxyruntimev1.EgressProfileSourceRef{
		SourceId: sourceSafeID(in.GetSourceId()),
		NodeId:   strings.TrimSpace(in.GetNodeId()),
	}
}

func normalizeEgressProfile(profile *proxyruntimev1.EgressProfileSettings) {
	if profile == nil {
		return
	}
	profile.ProfileId = sourceSafeID(profile.GetProfileId())
	profile.DisplayName = strings.TrimSpace(profile.GetDisplayName())
	profile.Line = egressProfileLineFromProto(profile.GetLine())
	profile.Exit = egressProfileExitFromProto(profile.GetExit())
}

func validateEgressProfile(profile *proxyruntimev1.EgressProfileSettings, index int, sourceIDs map[string]struct{}, dynamicProviderIDs map[string]struct{}) error {
	if profile.GetProfileId() == "" {
		return fmt.Errorf("egress_profiles[%d].profile_id is required", index)
	}
	if err := validateEgressProfileLine(profile.GetLine(), index, sourceIDs); err != nil {
		return err
	}
	return validateEgressProfileExit(profile.GetLine(), profile.GetExit(), index, sourceIDs, dynamicProviderIDs)
}

func validateEgressProfileLine(line *proxyruntimev1.EgressProfileLineSettings, index int, sourceIDs map[string]struct{}) error {
	switch line.GetKind() {
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_DIRECT:
		return nil
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_SOURCE:
		return validateEgressProfileSource(line.GetSource(), fmt.Sprintf("egress_profiles[%d].line.source", index), sourceIDs, true)
	default:
		return fmt.Errorf("egress_profiles[%d].line.kind is required", index)
	}
}

func validateEgressProfileExit(line *proxyruntimev1.EgressProfileLineSettings, exit *proxyruntimev1.EgressProfileExitSettings, index int, sourceIDs map[string]struct{}, dynamicProviderIDs map[string]struct{}) error {
	switch exit.GetKind() {
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DIRECT:
		return nil
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
		if err := validateEgressProfileSource(exit.GetSource(), fmt.Sprintf("egress_profiles[%d].exit.source", index), sourceIDs, true); err != nil {
			return err
		}
		if line.GetKind() == proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_SOURCE && sameProfileSource(line.GetSource(), exit.GetSource()) {
			return fmt.Errorf("egress_profiles[%d].line and exit cannot use the same source", index)
		}
		return nil
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
		if providerID := strings.TrimSpace(exit.GetDynamicProviderId()); providerID != "" && dynamicProviderIDs != nil {
			if _, exists := dynamicProviderIDs[providerID]; !exists {
				return fmt.Errorf("egress_profiles[%d].exit.dynamic_provider_id %q is not enabled", index, providerID)
			}
		}
		return nil
	default:
		return fmt.Errorf("egress_profiles[%d].exit.kind is required", index)
	}
}

func validateEgressProfileSource(source *proxyruntimev1.EgressProfileSourceRef, field string, sourceIDs map[string]struct{}, requireNode bool) error {
	sourceID := source.GetSourceId()
	if sourceID == "" {
		return fmt.Errorf("%s.source_id is required", field)
	}
	if requireNode && strings.TrimSpace(source.GetNodeId()) == "" {
		return fmt.Errorf("%s.node_id is required", field)
	}
	if sourceIDs != nil {
		if _, exists := sourceIDs[sourceID]; !exists {
			return fmt.Errorf("%s.source_id %q is not enabled", field, sourceID)
		}
	}
	return nil
}

func sameProfileSource(left *proxyruntimev1.EgressProfileSourceRef, right *proxyruntimev1.EgressProfileSourceRef) bool {
	return left.GetSourceId() == right.GetSourceId() && strings.TrimSpace(left.GetNodeId()) == strings.TrimSpace(right.GetNodeId())
}

func cloneEgressProfile(in *proxyruntimev1.EgressProfileSettings) *proxyruntimev1.EgressProfileSettings {
	return egressProfileFromProto(in)
}

func egressProfilesFromRequest(in []*proxyruntimev1.EgressProfileSettings, sourceIDs map[string]struct{}, dynamicProviderIDs map[string]struct{}) ([]*proxyruntimev1.EgressProfileSettings, error) {
	out := make([]*proxyruntimev1.EgressProfileSettings, 0, len(in))
	seen := map[string]struct{}{}
	for index, profile := range in {
		item := egressProfileFromProto(profile)
		if err := validateEgressProfile(item, index, sourceIDs, dynamicProviderIDs); err != nil {
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

func (s *runtimeSettingsStore) updateEgressProfiles(ctx context.Context, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	sourceIDs, err := s.enabledSourceIDs(ctx)
	if err != nil {
		return nil, err
	}
	dynamicProviderIDs := enabledDynamicProviderIDs(settings)
	nextProfiles, err := egressProfilesFromRequest(profiles, sourceIDs, dynamicProviderIDs)
	if err != nil {
		return nil, err
	}
	if err := rejectMissingIngressRuleProfiles(settings.GetIngressRules(), nextProfiles); err != nil {
		return nil, err
	}
	settings.EgressProfiles = nextProfiles
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) enabledSourceIDs(ctx context.Context) (map[string]struct{}, error) {
	if s.store == nil {
		return nil, nil
	}
	providers, fixedProxies, err := s.store.ListSourcePlaneConfig(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	for _, provider := range providers {
		if id := sourceSafeID(provider.ID); id != "" {
			out[id] = struct{}{}
		}
	}
	for _, fixed := range fixedProxies {
		if id := sourceSafeID(fixed.ID); id != "" {
			out[id] = struct{}{}
		}
	}
	return out, nil
}

func enabledDynamicProviderIDs(settings *runtimeSettingsFile) map[string]struct{} {
	out := map[string]struct{}{}
	for _, provider := range normalizeRuntimeSettings(settings).GetDynamicIpProviders() {
		if id := strings.TrimSpace(provider.GetProviderId()); id != "" {
			out[id] = struct{}{}
		}
	}
	return out
}

func enabledEgressProfileIDsFromProfiles(profiles []*proxyruntimev1.EgressProfileSettings) map[string]struct{} {
	out := map[string]struct{}{}
	for _, profile := range profiles {
		if id := sourceSafeID(profile.GetProfileId()); id != "" && profile.GetEnabled() {
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
		SourceID:       line.GetSource().GetSourceId(),
		NodeID:         strings.TrimSpace(line.GetSource().GetNodeId()),
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
		SourceID:       exit.GetSource().GetSourceId(),
		NodeID:         strings.TrimSpace(exit.GetSource().GetNodeId()),
		HealthCheckURL: strings.TrimSpace(exit.GetHealthCheckUrl()),
		HealthInterval: protoDuration(exit.GetHealthInterval(), 300*time.Second),
		HealthTimeout:  protoDuration(exit.GetHealthTimeout(), 5*time.Second),
		ExpectedStatus: defaultExpectedStatus(exit.GetExpectedStatus()),
	}
}

func egressProfileLineKind(kind proxyruntimev1.EgressProfileLineKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_SOURCE:
		return "source"
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
