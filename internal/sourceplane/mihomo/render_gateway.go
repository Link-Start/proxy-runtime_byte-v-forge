package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

const gatewayListenerName = "proxy-runtime-gateway"

func renderGateway(endpoint sourceplane.Endpoint, users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute) (mihomoListener, []mihomoGroup, []string, error) {
	host, port, err := splitEndpoint(endpoint.Addr)
	if err != nil {
		return mihomoListener{}, nil, nil, err
	}
	listener := mihomoListener{Name: gatewayListenerName, Type: "mixed", Listen: host, Port: port, UDP: true}
	listener.Users = renderUsers(users, sessions)
	groups := renderUserGroups(users, sessions)
	rules := renderUserRules(users, sessions)
	return listener, groups, rules, nil
}

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

func renderUserRules(users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute) []string {
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
		rules = append(rules, "IN-USER,"+username+","+userRouteTarget(user))
	}
	return rules
}

func renderUserGroups(users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute) []mihomoGroup {
	groups := make([]mihomoGroup, 0, len(users)+len(sessions))
	for _, session := range sessions {
		if len(session.Pool) == 0 {
			continue
		}
		name := sessionGroupName(session)
		if name == "" {
			continue
		}
		groups = append(groups, mihomoGroup{Name: name, Type: "select", Proxies: sessionProxyNames(session), Hidden: true})
	}
	for _, user := range users {
		target := userRouteTarget(user)
		if target == groupName || target == "DIRECT" {
			continue
		}
		if strings.TrimSpace(user.Route) == "source" {
			if group := sourceRouteGroup(target, user); group != nil {
				groups = append(groups, *group)
			}
		}
	}
	return groups
}

func sourceRouteGroup(name string, user dataplane.ProxyUserRoute) *mihomoGroup {
	sourceID := safeID(user.SourceID)
	if sourceID == "" {
		return nil
	}
	group := &mihomoGroup{Name: name, Type: "select", Hidden: true}
	group.Proxies = []string{safeID(firstNonEmpty(user.NodeID, user.SourceID))}
	group.Use = []string{sourceID}
	return group
}

func userRouteTarget(user dataplane.ProxyUserRoute) string {
	switch strings.TrimSpace(user.Route) {
	case "direct":
		return "DIRECT"
	case "source":
		return userGroupName(user)
	case "profile":
		return profileGroupName(firstNonEmpty(user.ProfileID, user.SourceID))
	default:
		return groupName
	}
}

func userGroupName(user dataplane.ProxyUserRoute) string {
	id := safeID(firstNonEmpty(user.ID, user.Username))
	if id == "" {
		id = "proxy-user"
	}
	return "bvf-user-" + id
}

func sessionGroupName(session dataplane.SessionRoute) string {
	id := safeID(session.SessionID)
	if id == "" {
		return ""
	}
	return "bvf-session-" + id
}

func sessionProxyNames(session dataplane.SessionRoute) []string {
	out := make([]string, 0, len(session.Pool))
	for index, node := range session.Pool {
		id := safeID(firstNonEmpty(node.ID, fmt.Sprintf("session-%d", index)))
		if id == "" {
			id = fmt.Sprintf("session-%d", index)
		}
		out = append(out, sessionProxyName(session, id, index))
	}
	return out
}

func sessionProxyName(session dataplane.SessionRoute, nodeID string, index int) string {
	sessionID := safeID(firstNonEmpty(session.SessionID, "session"))
	return fmt.Sprintf("%s-%s-%d", sessionID, nodeID, index)
}
