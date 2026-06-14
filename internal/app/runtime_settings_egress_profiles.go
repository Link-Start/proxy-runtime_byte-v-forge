package app

import (
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

func cloneDuration(value *durationpb.Duration) *durationpb.Duration {
	if value == nil {
		return nil
	}
	return durationpb.New(value.AsDuration())
}
