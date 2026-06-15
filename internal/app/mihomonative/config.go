package mihomonative

const (
	fileName                    = "native.json"
	fixedProxyGroupName         = "固定代理"
	defaultProviderHealthURL    = "https://www.gstatic.com/generate_204"
	defaultProviderHealthPeriod = 300
	defaultProviderHealthWait   = 5000
	defaultProviderUserAgent    = "mihomo/1.18.3"
)

type FixedProxy struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	URI  string `json:"uri"`
}

type Subscription struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ConfigFile struct {
	FixedProxies   []FixedProxy        `json:"fixed_proxies,omitempty"`
	Proxies        []map[string]any    `json:"proxies,omitempty"`
	Subscriptions  []Subscription      `json:"subscriptions,omitempty"`
	ProxyProviders map[string]Provider `json:"proxy-providers,omitempty"`
	ProxyGroups    []Group             `json:"proxy-groups,omitempty"`
	Rules          []string            `json:"rules,omitempty"`
}

type Provider struct {
	Type        string              `json:"type"`
	URL         string              `json:"url,omitempty"`
	Path        string              `json:"path,omitempty"`
	Interval    int                 `json:"interval,omitempty"`
	Filter      string              `json:"filter,omitempty"`
	Exclude     string              `json:"exclude-filter,omitempty"`
	HealthCheck *HealthCheck        `json:"health-check,omitempty"`
	Header      map[string][]string `json:"header,omitempty"`
	Override    map[string]any      `json:"override,omitempty"`
}

type HealthCheck struct {
	Enable         bool   `json:"enable"`
	URL            string `json:"url,omitempty"`
	Interval       int    `json:"interval,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	Lazy           bool   `json:"lazy"`
	ExpectedStatus uint32 `json:"expected-status,omitempty"`
}

type Group struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Proxies        []string `json:"proxies,omitempty"`
	Use            []string `json:"use,omitempty"`
	Filter         string   `json:"filter,omitempty"`
	URL            string   `json:"url,omitempty"`
	Interval       int      `json:"interval,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	Strategy       string   `json:"strategy,omitempty"`
	Lazy           bool     `json:"lazy"`
	ExpectedStatus uint32   `json:"expected-status,omitempty"`
	Hidden         bool     `json:"hidden,omitempty"`
}
