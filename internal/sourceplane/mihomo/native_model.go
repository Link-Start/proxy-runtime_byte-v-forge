package mihomo

type mihomoNativeConfig struct {
	FixedProxies   []mihomoNativeFixedProxy   `json:"fixed_proxies,omitempty"`
	Proxies        []map[string]any           `json:"proxies,omitempty"`
	Subscriptions  []mihomoNativeSubscription `json:"subscriptions,omitempty"`
	ProxyProviders map[string]mihomoProvider  `json:"proxy-providers,omitempty"`
	ProxyGroups    []mihomoGroup              `json:"proxy-groups,omitempty"`
	Rules          []string                   `json:"rules,omitempty"`
}

type mihomoNativeFixedProxy struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type mihomoNativeSubscription struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}
