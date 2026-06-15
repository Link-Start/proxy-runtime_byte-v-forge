package mihomonative

import (
	"strings"
)

func preserveGroups(groups []Group) []Group {
	out := make([]Group, 0, len(groups))
	for _, group := range groups {
		if strings.TrimSpace(group.Name) == "" || group.Name == fixedProxyGroupName {
			continue
		}
		out = append(out, group)
	}
	return out
}
