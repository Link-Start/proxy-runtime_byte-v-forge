package mihomo

type mihomoConfig struct {
	MixedPort          int                       `json:"mixed-port,omitempty"`
	BindAddress        string                    `json:"bind-address,omitempty"`
	AllowLAN           bool                      `json:"allow-lan"`
	Mode               string                    `json:"mode"`
	LogLevel           string                    `json:"log-level"`
	ExternalController string                    `json:"external-controller,omitempty"`
	Secret             string                    `json:"secret,omitempty"`
	ExternalUI         string                    `json:"external-ui,omitempty"`
	ExternalUIURL      string                    `json:"external-ui-url,omitempty"`
	Authentication     []string                  `json:"authentication,omitempty"`
	Proxies            []map[string]any          `json:"proxies,omitempty"`
	ProxyProviders     map[string]mihomoProvider `json:"proxy-providers,omitempty"`
	ProxyGroups        []mihomoGroup             `json:"proxy-groups,omitempty"`
	Listeners          []mihomoListener          `json:"listeners,omitempty"`
	Rules              []string                  `json:"rules"`
}

type mihomoGroup struct {
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
