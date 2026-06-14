package mihomo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func profileLayerGroup(name string, layer sourceplane.EgressProfileLayer, groupType string) mihomoGroup {
	return mihomoGroup{
		Name:           name,
		Type:           profileGroupStrategy(groupType),
		URL:            firstNonEmpty(layer.HealthCheckURL, "https://www.gstatic.com/generate_204"),
		Interval:       seconds(layer.HealthInterval, 300),
		Timeout:        milliseconds(layer.HealthTimeout, 5000),
		Lazy:           true,
		ExpectedStatus: defaultExpectedStatus(layer.ExpectedStatus),
		Hidden:         true,
	}
}

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

func mihomoRuleTargetName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var out strings.Builder
	lastSpace := false
	for _, r := range value {
		if r == ',' || r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		if r == ' ' {
			if lastSpace {
				continue
			}
			lastSpace = true
			out.WriteRune(r)
			continue
		}
		lastSpace = false
		out.WriteRune(r)
	}
	return strings.TrimSpace(out.String())
}

func exactNodeFilter(name string) string {
	return "^" + regexp.QuoteMeta(strings.TrimSpace(name)) + "$"
}

func profileGroupStrategy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "url-test", "select":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "select"
	}
}
