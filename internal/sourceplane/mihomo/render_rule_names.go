package mihomo

import (
	"regexp"
	"strings"
)

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
