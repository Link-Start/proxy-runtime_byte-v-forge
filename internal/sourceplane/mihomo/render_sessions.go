package mihomo

import (
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

func renderSessionRoutes(sessions []dataplane.SessionRoute) ([]map[string]any, error) {
	out := make([]map[string]any, 0)
	for _, session := range sessions {
		for index, node := range session.Pool {
			id := safeID(firstNonEmpty(node.ID, fmt.Sprintf("session-%d", index)))
			if id == "" {
				id = fmt.Sprintf("session-%d", index)
			}
			name := sessionProxyName(session, id, index)
			proxy, err := renderProxyURL(name, node.URL)
			if err != nil {
				return nil, err
			}
			out = append(out, proxy)
		}
	}
	return out, nil
}
