package app

import "strings"

type mihomoConnectionSelector struct {
	inboundUsers []string
	chains       []string
}

func connectionMatches(connection mihomoConnection, inboundUsers map[string]struct{}, chains map[string]struct{}) bool {
	if _, ok := inboundUsers[normalizedKey(connection.Metadata.InboundUser)]; ok {
		return true
	}
	if connection.Rule == "InUser" {
		if _, ok := inboundUsers[normalizedKey(connection.RulePayload)]; ok {
			return true
		}
	}
	for _, chain := range connection.Chains {
		if _, ok := chains[normalizedKey(chain)]; ok {
			return true
		}
	}
	return false
}

func normalizedSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		if key := normalizedKey(value); key != "" {
			out[key] = struct{}{}
		}
	}
	return out
}

func normalizedKey(value string) string {
	return strings.TrimSpace(value)
}
