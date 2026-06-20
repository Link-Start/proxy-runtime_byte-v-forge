package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func dynamicLeaseLineProfiles(settings *runtimeSettingsFile, profileID string) []*proxyruntimev1.EgressProfileSettings {
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
		if profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		line := profile.GetLine()
		if line.GetKind() != proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE {
			continue
		}
		return []*proxyruntimev1.EgressProfileSettings{profile}
	}
	return nil
}
