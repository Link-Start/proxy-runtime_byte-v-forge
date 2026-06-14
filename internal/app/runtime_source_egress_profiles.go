package app

import "github.com/byte-v-forge/proxy-runtime/internal/sourceplane"

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
