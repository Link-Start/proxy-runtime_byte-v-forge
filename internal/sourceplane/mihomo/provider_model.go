package mihomo

type mihomoProvider struct {
	Type        string              `json:"type"`
	URL         string              `json:"url,omitempty"`
	Path        string              `json:"path,omitempty"`
	Proxy       string              `json:"proxy,omitempty"`
	Interval    int                 `json:"interval,omitempty"`
	Filter      string              `json:"filter,omitempty"`
	Exclude     string              `json:"exclude-filter,omitempty"`
	HealthCheck *mihomoHealthCheck  `json:"health-check,omitempty"`
	Header      map[string][]string `json:"header,omitempty"`
	Override    map[string]any      `json:"override,omitempty"`
}

type mihomoHealthCheck struct {
	Enable         bool   `json:"enable"`
	URL            string `json:"url,omitempty"`
	Interval       int    `json:"interval,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	Lazy           bool   `json:"lazy"`
	ExpectedStatus uint32 `json:"expected-status,omitempty"`
}
