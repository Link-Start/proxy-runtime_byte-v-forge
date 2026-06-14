package mihomo

func appendUniqueGroups(base []mihomoGroup, groups ...[]mihomoGroup) []mihomoGroup {
	out := make([]mihomoGroup, 0, len(base))
	seen := map[string]struct{}{}
	for _, group := range base {
		if group.Name == "" {
			continue
		}
		seen[group.Name] = struct{}{}
		out = append(out, group)
	}
	for _, items := range groups {
		for _, group := range items {
			if group.Name == "" {
				continue
			}
			if _, exists := seen[group.Name]; exists {
				continue
			}
			seen[group.Name] = struct{}{}
			out = append(out, group)
		}
	}
	return out
}
