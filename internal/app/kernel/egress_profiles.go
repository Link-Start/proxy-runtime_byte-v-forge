package kernel

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func NormalizeEgressProfile(profile *proxygatewayv1.EgressProfileSettings) {
	if profile == nil {
		return
	}
	profile.ProfileId = appcore.RuntimeSafeID(profile.GetProfileId())
	profile.DisplayName = strings.TrimSpace(profile.GetDisplayName())
	profile.Line = EgressProfileLineFromProto(profile.GetLine())
	profile.Exit = EgressProfileExitFromProto(profile.GetExit())
}

func EgressProfileLineFromProto(in *proxygatewayv1.EgressProfileLineSettings) *proxygatewayv1.EgressProfileLineSettings {
	if in == nil {
		in = &proxygatewayv1.EgressProfileLineSettings{}
	}
	out := &proxygatewayv1.EgressProfileLineSettings{
		Kind:           in.GetKind(),
		MihomoNode:     EgressProfileMihomoNodeRefFromProto(in.GetMihomoNode()),
		HealthCheckUrl: strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval: appcore.CloneDuration(in.GetHealthInterval()),
		HealthTimeout:  appcore.CloneDuration(in.GetHealthTimeout()),
		ExpectedStatus: in.GetExpectedStatus(),
	}
	if out.GetKind() == proxygatewayv1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_UNSPECIFIED {
		out.Kind = proxygatewayv1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_DIRECT
	}
	return out
}

func EgressProfileExitFromProto(in *proxygatewayv1.EgressProfileExitSettings) *proxygatewayv1.EgressProfileExitSettings {
	if in == nil {
		in = &proxygatewayv1.EgressProfileExitSettings{}
	}
	out := &proxygatewayv1.EgressProfileExitSettings{
		Kind:              in.GetKind(),
		MihomoNode:        EgressProfileMihomoNodeRefFromProto(in.GetMihomoNode()),
		HealthCheckUrl:    strings.TrimSpace(in.GetHealthCheckUrl()),
		HealthInterval:    appcore.CloneDuration(in.GetHealthInterval()),
		HealthTimeout:     appcore.CloneDuration(in.GetHealthTimeout()),
		ExpectedStatus:    in.GetExpectedStatus(),
		DynamicProviderId: strings.TrimSpace(in.GetDynamicProviderId()),
		DynamicIpPolicy:   EgressProfileDynamicIPPolicyFromProto(in.GetDynamicIpPolicy()),
	}
	return out
}

func EgressProfileDynamicIPPolicyFromProto(in *proxygatewayv1.ProxySessionPolicy) *proxygatewayv1.ProxySessionPolicy {
	if in == nil {
		return nil
	}
	return NormalizeDynamicIPSessionPolicy(in)
}

func EgressProfileMihomoNodeRefFromProto(in *proxygatewayv1.EgressProfileMihomoNodeRef) *proxygatewayv1.EgressProfileMihomoNodeRef {
	if in == nil {
		return &proxygatewayv1.EgressProfileMihomoNodeRef{}
	}
	return &proxygatewayv1.EgressProfileMihomoNodeRef{
		ResourceId: strings.TrimSpace(in.GetResourceId()),
		NodeId:     strings.TrimSpace(in.GetNodeId()),
	}
}
