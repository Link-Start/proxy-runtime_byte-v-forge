package config

import (
	"errors"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider/ten24"
)

const (
	ProviderTen24  = "1024proxy"
	ProviderB2     = "b2proxy"
	ProviderCli    = "cliproxy"
	ProviderNone   = "none"
	ProviderStatic = "static"
)

var ErrUnsupportedProvider = errors.New("unsupported proxy provider")

const (
	ListenerRouteDirect   = "direct"
	ListenerRouteProvider = "provider"
	ListenerRouteProfile  = "profile"
)

type EgressListener struct {
	ID       string            `json:"id"`
	Addr     string            `json:"addr"`
	Protocol string            `json:"protocol"`
	Route    string            `json:"route"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Labels   map[string]string `json:"labels"`
}

type ProxyUserRoute struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Route     string `json:"route"`
	SourceID  string `json:"source_id"`
	NodeID    string `json:"node_id"`
	ProfileID string `json:"profile_id"`
}

type SessionListenerConfig struct {
	AdvertisedHost string
}

type MihomoConfig struct {
	Path                string
	ConfigDir           string
	APIAddr             string
	DashboardDir        string
	DashboardURL        string
	GroupStrategy       string
	HealthCheckURL      string
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type IPFraudConfig struct {
	Timeout     time.Duration
	CacheTTL    time.Duration
	KeyCooldown time.Duration
}

type Config struct {
	RuntimeAddr       string
	PostgresDSN       string
	RedisURL          string
	EncryptionKey     string
	Mihomo            MihomoConfig
	CommonEgressAddr  string
	LocalAddr         string
	LocalProtocol     string
	LocalUsername     string
	LocalPassword     string
	SessionListener   SessionListenerConfig
	ProxyUsers        []ProxyUserRoute
	SimpleProxies     []string
	ProviderHTTPProxy string
	Provider          string
	RefreshInterval   time.Duration
	RequestTimeout    time.Duration
	ProxyExitGeoURLs  []string
	IPFraud           IPFraudConfig
	EdgeCanaryTimeout time.Duration
	Ten24             ten24.Config
}
