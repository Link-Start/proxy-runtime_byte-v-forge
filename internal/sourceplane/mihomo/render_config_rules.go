package mihomo

func renderConfigRules(userRules []string, nativeRules []string) []string {
	rules := append(userRules, nativeRules...)
	return append(rules, "MATCH,REJECT")
}

func baseMihomoGroups() []mihomoGroup {
	return []mihomoGroup{
		{
			Name:    "GLOBAL",
			Type:    "select",
			Proxies: []string{"REJECT"},
			Hidden:  true,
		},
	}
}
