package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func validateProxyUsers(users []ProxyUserRoute) error {
	seen := map[string]struct{}{}
	for index, user := range users {
		username := strings.TrimSpace(user.Username)
		if username == "" {
			return fmt.Errorf("PROXY_RUNTIME_PROXY_USERS_JSON[%d].username is required", index)
		}
		if _, ok := seen[username]; ok {
			return fmt.Errorf("duplicate proxy user username %q", username)
		}
		seen[username] = struct{}{}
		switch normalizeConfigToken(user.Route) {
		case "", ListenerRouteProvider, ListenerRouteDirect, ListenerRouteProfile, "source":
		default:
			return fmt.Errorf("unsupported proxy user route %q", user.Route)
		}
		if normalizeConfigToken(user.Route) == ListenerRouteProfile && strings.TrimSpace(user.ProfileID) == "" && strings.TrimSpace(user.SourceID) == "" {
			return fmt.Errorf("PROXY_RUNTIME_PROXY_USERS_JSON[%d].profile_id is required for profile route", index)
		}
	}
	return nil
}

func envProxyUsers(name string) []ProxyUserRoute {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}
	var users []ProxyUserRoute
	if err := json.Unmarshal([]byte(raw), &users); err != nil {
		return []ProxyUserRoute{{
			Username: "__invalid__",
			Route:    fmt.Sprintf("invalid JSON: %v", err),
		}}
	}
	for index := range users {
		users[index].ID = strings.TrimSpace(users[index].ID)
		users[index].Username = strings.TrimSpace(users[index].Username)
		users[index].Password = strings.TrimSpace(users[index].Password)
		users[index].Route = normalizeConfigToken(users[index].Route)
		users[index].SourceID = strings.TrimSpace(users[index].SourceID)
		users[index].NodeID = strings.TrimSpace(users[index].NodeID)
		users[index].ProfileID = strings.TrimSpace(users[index].ProfileID)
	}
	return users
}

func isLocalProtocol(protocol string) bool {
	switch protocol {
	case "http", "socks5":
		return true
	default:
		return false
	}
}
