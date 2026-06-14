package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func dynamicLeaseLineProfiles(settings *runtimeSettingsFile, profileID string) []*proxyruntimev1.EgressProfileSettings {
	profileID = runtimeSafeID(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		if runtimeSafeID(profile.GetProfileId()) != profileID {
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
