package mihomo

import (
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

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
	return groups
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
