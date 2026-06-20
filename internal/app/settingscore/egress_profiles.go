package settingscore

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func NormalizeEgressProfile(profile *proxyruntimev1.EgressProfileSettings) {
	if profile == nil {
		return
	}
	profile.ProfileId = appcore.RuntimeSafeID(profile.GetProfileId())
	profile.DisplayName = strings.TrimSpace(profile.GetDisplayName())
	profile.Line = EgressProfileLineFromProto(profile.GetLine())
	profile.Exit = EgressProfileExitFromProto(profile.GetExit())
}

func EgressProfileLineFromProto(in *proxyruntimev1.EgressProfileLineSettings) *proxyruntimev1.EgressProfileLineSettings {
	if in == nil {
		in = &proxyruntimev1.EgressProfileLineSettings{}
	}
	out := &proxyruntimev1.EgressProfileLineSettings{
		Kind:           in.GetKind(),
		MihomoNode:     EgressProfileMihomoNodeRefFromProto(in.GetMihomoNode()),
		HealthCheckUrl: strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval: CloneDuration(in.GetHealthInterval()),
		HealthTimeout:  CloneDuration(in.GetHealthTimeout()),
		ExpectedStatus: in.GetExpectedStatus(),
	}
	if out.GetKind() == proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_UNSPECIFIED {
		out.Kind = proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_DIRECT
	}
	return out
}

func EgressProfileExitFromProto(in *proxyruntimev1.EgressProfileExitSettings) *proxyruntimev1.EgressProfileExitSettings {
	if in == nil {
		in = &proxyruntimev1.EgressProfileExitSettings{}
	}
	out := &proxyruntimev1.EgressProfileExitSettings{
		Kind:              in.GetKind(),
		MihomoNode:        EgressProfileMihomoNodeRefFromProto(in.GetMihomoNode()),
		HealthCheckUrl:    strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval:    CloneDuration(in.GetHealthInterval()),
		HealthTimeout:     CloneDuration(in.GetHealthTimeout()),
		ExpectedStatus:    in.GetExpectedStatus(),
		DynamicProviderId: strings.TrimSpace(in.GetDynamicProviderId()),
		DynamicIpPolicy:   EgressProfileDynamicIPPolicyFromProto(in.GetDynamicIpPolicy()),
	}
	return out
}

func EgressProfileDynamicIPPolicyFromProto(in *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	if in == nil {
		return nil
	}
	return leaseapp.NormalizeDynamicIPSessionPolicy(in)
}

func EgressProfileMihomoNodeRefFromProto(in *proxyruntimev1.EgressProfileMihomoNodeRef) *proxyruntimev1.EgressProfileMihomoNodeRef {
	if in == nil {
		return &proxyruntimev1.EgressProfileMihomoNodeRef{}
	}
	return &proxyruntimev1.EgressProfileMihomoNodeRef{
		ResourceId: strings.TrimSpace(in.GetResourceId()),
		NodeId:     strings.TrimSpace(in.GetNodeId()),
	}
}

func CloneDuration(value *durationpb.Duration) *durationpb.Duration {
	if value == nil {
		return nil
	}
	return durationpb.New(value.AsDuration())
}
