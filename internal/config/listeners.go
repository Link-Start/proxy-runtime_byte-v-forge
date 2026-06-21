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
			return fmt.Errorf("PROXY_GATEWAY_PROXY_USERS_JSON[%d].username is required", index)
		}
		if _, ok := seen[username]; ok {
			return fmt.Errorf("duplicate proxy user username %q", username)
		}
		seen[username] = struct{}{}
		switch normalizeConfigToken(user.Route) {
		case "", ListenerRouteDirect, ListenerRouteProfile:
		default:
			return fmt.Errorf("unsupported proxy user route %q", user.Route)
		}
		if normalizeConfigToken(user.Route) == ListenerRouteProfile && strings.TrimSpace(user.ProfileID) == "" {
			return fmt.Errorf("PROXY_GATEWAY_PROXY_USERS_JSON[%d].profile_id is required for profile route", index)
		}
	}
	return nil
}

func envProxyUsers(name string) ([]ProxyUserRoute, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil, nil
	}
	var users []ProxyUserRoute
	if err := json.Unmarshal([]byte(raw), &users); err != nil {
		return nil, fmt.Errorf("%s must be valid JSON array: %w", name, err)
	}
	for index := range users {
		users[index].ID = strings.TrimSpace(users[index].ID)
		users[index].Username = strings.TrimSpace(users[index].Username)
		users[index].Password = strings.TrimSpace(users[index].Password)
		users[index].Route = normalizeConfigToken(users[index].Route)
		users[index].ProfileID = strings.TrimSpace(users[index].ProfileID)
	}
	return users, nil
}

func isLocalProtocol(protocol string) bool {
	switch protocol {
	case "http", "socks5":
		return true
	default:
		return false
	}
}
