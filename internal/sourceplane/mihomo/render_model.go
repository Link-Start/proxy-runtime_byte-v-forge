package mihomo

import (
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type renderOptions struct {
	EgressProfiles      []sourceplane.EgressProfile
	Endpoint            sourceplane.Endpoint
	ConfigDir           string
	NativeConfig        mihomoNativeConfig
	APIAddr             string
	DashboardDir        string
	DashboardURL        string
	HealthCheckURL      string
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	BasePool            []provider.Node
	AvailableProxies    map[string]struct{}
	AvailableProviders  map[string]struct{}
	ProxyUsers          []dataplane.ProxyUserRoute
	SessionRoutes       []dataplane.SessionRoute
	ProfileGroups       map[string]string
}

type mihomoConfig struct {
	MixedPort          int                       `json:"mixed-port,omitempty"`
	BindAddress        string                    `json:"bind-address,omitempty"`
	AllowLAN           bool                      `json:"allow-lan"`
	Mode               string                    `json:"mode"`
	LogLevel           string                    `json:"log-level"`
	ExternalController string                    `json:"external-controller,omitempty"`
	ExternalUI         string                    `json:"external-ui,omitempty"`
	ExternalUIURL      string                    `json:"external-ui-url,omitempty"`
	Authentication     []string                  `json:"authentication,omitempty"`
	Proxies            []map[string]any          `json:"proxies,omitempty"`
	ProxyProviders     map[string]mihomoProvider `json:"proxy-providers,omitempty"`
	ProxyGroups        []mihomoGroup             `json:"proxy-groups,omitempty"`
	Listeners          []mihomoListener          `json:"listeners,omitempty"`
	Rules              []string                  `json:"rules"`
}

type mihomoNativeConfig struct {
	Proxies        []map[string]any          `json:"proxies,omitempty"`
	ProxyProviders map[string]mihomoProvider `json:"proxy-providers,omitempty"`
	ProxyGroups    []mihomoGroup             `json:"proxy-groups,omitempty"`
	Rules          []string                  `json:"rules,omitempty"`
}

type mihomoProvider struct {
	Type        string              `json:"type"`
	URL         string              `json:"url"`
	Path        string              `json:"path,omitempty"`
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

type mihomoListener struct {
	Name   string       `json:"name"`
	Type   string       `json:"type"`
	Listen string       `json:"listen,omitempty"`
	Port   int          `json:"port"`
	Rule   string       `json:"rule,omitempty"`
	Proxy  string       `json:"proxy,omitempty"`
	UDP    bool         `json:"udp"`
	Users  []mihomoUser `json:"users,omitempty"`
}

type mihomoUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
