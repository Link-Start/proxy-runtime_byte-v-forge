package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func egressProfileFromProto(in *proxyruntimev1.EgressProfileSettings) *proxyruntimev1.EgressProfileSettings {
	if in == nil {
		return &proxyruntimev1.EgressProfileSettings{}
	}
	out := &proxyruntimev1.EgressProfileSettings{
		ProfileId:   appcore.RuntimeSafeID(in.GetProfileId()),
		DisplayName: strings.TrimSpace(in.GetDisplayName()),
		Enabled:     in.GetEnabled(),
		Line:        settingscore.EgressProfileLineFromProto(in.GetLine()),
		Exit:        settingscore.EgressProfileExitFromProto(in.GetExit()),
	}
	settingscore.NormalizeEgressProfile(out)
	return out
}
