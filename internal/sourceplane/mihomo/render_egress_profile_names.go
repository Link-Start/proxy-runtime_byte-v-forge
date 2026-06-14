package mihomo

import (
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func profileInternalGroupName(profileID string) string {
	id := safeID(profileID)
	if id == "" {
		id = "profile"
	}
	return "bvf-profile-" + id
}

func profileLineGroupName(profileID string) string {
	return profileInternalGroupName(profileID) + "-line"
}

func profileGroupNames(profiles []sourceplane.EgressProfile) map[string]string {
	out := map[string]string{}
	used := map[string]int{}
	for _, profile := range profiles {
		if !profile.Enabled {
			continue
		}
		key := profileIDKey(profile.ID)
		if key == "" {
			continue
		}
		base := mihomoRuleTargetName(firstNonEmpty(profile.DisplayName, profile.ID))
		if base == "" {
			base = profileInternalGroupName(profile.ID)
		}
		name := base
		if count := used[base]; count > 0 {
			name = fmt.Sprintf("%s %d", base, count+1)
		}
		used[base]++
		out[key] = name
	}
	return out
}

func profileGroupNameFor(opts renderOptions, profile sourceplane.EgressProfile) string {
	if name := opts.ProfileGroups[profileIDKey(profile.ID)]; name != "" {
		return name
	}
	return profileInternalGroupName(profile.ID)
}

func profileIDKey(value string) string {
	return safeID(value)
}
