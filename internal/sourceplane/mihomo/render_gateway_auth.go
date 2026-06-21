package mihomo

import (
	"strings"

	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
)

func renderUsers(users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute) []mihomoUser {
	seen := map[string]struct{}{}
	out := make([]mihomoUser, 0, len(users)+len(sessions))
	for _, user := range users {
		username := strings.TrimSpace(user.Username)
		if username == "" {
			continue
		}
		if _, exists := seen[username]; exists {
			continue
		}
		seen[username] = struct{}{}
		out = append(out, mihomoUser{Username: username, Password: user.Password})
	}
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
		out = append(out, mihomoUser{Username: username, Password: session.Listener.Password})
	}
	return out
}

func renderAuthentication(users []mihomoUser) []string {
	out := make([]string, 0, len(users))
	for _, user := range users {
		username := strings.TrimSpace(user.Username)
		if username == "" {
			continue
		}
		out = append(out, username+":"+user.Password)
	}
	return out
}
