package domain

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func egressProfileFromProto(in *proxygatewayv1.EgressProfileSettings) *proxygatewayv1.EgressProfileSettings {
	if in == nil {
		return &proxygatewayv1.EgressProfileSettings{}
	}
	out := &proxygatewayv1.EgressProfileSettings{
		ProfileId:   appcore.RuntimeSafeID(in.GetProfileId()),
		DisplayName: strings.TrimSpace(in.GetDisplayName()),
		Enabled:     in.GetEnabled(),
		Line:        kernel.EgressProfileLineFromProto(in.GetLine()),
		Exit:        kernel.EgressProfileExitFromProto(in.GetExit()),
	}
	kernel.NormalizeEgressProfile(out)
	return out
}
