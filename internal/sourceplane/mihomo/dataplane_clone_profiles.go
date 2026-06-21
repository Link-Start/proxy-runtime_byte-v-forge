package mihomo

import "github.com/byte-v-forge/proxy-gateway/internal/sourceplane"

func cloneEgressProfiles(profiles []sourceplane.EgressProfile) []sourceplane.EgressProfile {
	if len(profiles) == 0 {
		return nil
	}
	out := make([]sourceplane.EgressProfile, len(profiles))
	for index, profile := range profiles {
		out[index] = profile
	}
	return out
}
