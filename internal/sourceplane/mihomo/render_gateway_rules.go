package mihomo

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

func renderUserRules(users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute, profiles map[string]string) []string {
	rules := make([]string, 0, len(users)+len(sessions))
	seen := map[string]struct{}{}
	for _, session := range sessions {
		if len(session.Pool) == 0 {
			continue
		}
		username := strings.TrimSpace(session.Listener.Username)
		if username == "" {
			continue
		}
		if _, exists := seen[username]; exists {
			continue
		}
		seen[username] = struct{}{}
		rules = append(rules, "IN-USER,"+username+","+sessionGroupName(session))
	}
	for _, user := range users {
		username := strings.TrimSpace(user.Username)
		if username == "" {
			continue
		}
		if _, exists := seen[username]; exists {
			continue
		}
		seen[username] = struct{}{}
		rules = append(rules, "IN-USER,"+username+","+userRouteTarget(user, profiles))
	}
	return rules
}

func userRouteTarget(user dataplane.ProxyUserRoute, profiles map[string]string) string {
	switch strings.TrimSpace(user.Route) {
	case "direct":
		return "DIRECT"
	case "profile":
		if name := profiles[profileIDKey(user.ProfileID)]; name != "" {
			return name
		}
		return profileInternalGroupName(user.ProfileID)
	default:
		return "REJECT"
	}
}
