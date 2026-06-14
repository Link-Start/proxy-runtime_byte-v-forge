package config

import (
	"os"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider/ten24"
)

func LoadFromEnv() (Config, error) {
	mihomoConfigDir := envStringDefault("PROXY_RUNTIME_MIHOMO_CONFIG_DIR", "/var/lib/proxy-runtime/mihomo")
	mihomoHealthCheckInterval, err := envDurationSeconds("PROXY_RUNTIME_MIHOMO_HEALTH_CHECK_INTERVAL_SECONDS", 300*time.Second)
	if err != nil {
		return Config{}, err
	}
	mihomoHealthCheckTimeout, err := envDurationSeconds("PROXY_RUNTIME_MIHOMO_HEALTH_CHECK_TIMEOUT_SECONDS", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	proxyUsers, err := envProxyUsers("PROXY_RUNTIME_PROXY_USERS_JSON")
	if err != nil {
		return Config{}, err
	}
	refreshInterval, err := envDurationSeconds("PROXY_RUNTIME_REFRESH_SECONDS", 300*time.Second)
	if err != nil {
		return Config{}, err
	}
	requestTimeout, err := envDurationSeconds("PROXY_RUNTIME_REQUEST_TIMEOUT_SECONDS", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	edgeCanaryTimeout, err := envDurationSeconds("PROXY_RUNTIME_EDGE_CANARY_TIMEOUT_SECONDS", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	ipFraudTimeout, err := envDurationSeconds("PROXY_RUNTIME_IP_FRAUD_TIMEOUT_SECONDS", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	ipFraudCacheTTL, err := envDurationSeconds("PROXY_RUNTIME_IP_FRAUD_CACHE_TTL_SECONDS", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}
	ipFraudKeyCooldown, err := envDurationSeconds("PROXY_RUNTIME_IP_FRAUD_KEY_COOLDOWN_SECONDS", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		RuntimeAddr:   envStringDefault("PROXY_RUNTIME_ADDR", ":8080"),
		DataDir:       envStringDefault("PROXY_RUNTIME_DATA_DIR", "/var/lib/proxy-runtime"),
		PostgresDSN:   firstNonEmpty(strings.TrimSpace(os.Getenv("PROXY_RUNTIME_POSTGRES_DSN")), strings.TrimSpace(os.Getenv("PG_DSN"))),
		RedisURL:      strings.TrimSpace(os.Getenv("PROXY_RUNTIME_REDIS_URL")),
		EncryptionKey: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_ENCRYPTION_KEY")),
		Mihomo: MihomoConfig{
			Path:                envStringDefault("PROXY_RUNTIME_MIHOMO_PATH", "mihomo"),
			ConfigDir:           mihomoConfigDir,
			APIAddr:             envStringDefault("PROXY_RUNTIME_MIHOMO_API_ADDR", "127.0.0.1:18901"),
			DashboardDir:        envStringDefault("PROXY_RUNTIME_MIHOMO_DASHBOARD_DIR", "/app/dashboard/metacubexd"),
			DashboardURL:        strings.TrimSpace(os.Getenv("PROXY_RUNTIME_MIHOMO_DASHBOARD_URL")),
			HealthCheckURL:      envStringDefault("PROXY_RUNTIME_MIHOMO_HEALTH_CHECK_URL", "https://www.gstatic.com/generate_204"),
			HealthCheckInterval: mihomoHealthCheckInterval,
			HealthCheckTimeout:  mihomoHealthCheckTimeout,
		},
		LocalAddr:     envStringDefault("PROXY_RUNTIME_DYNAMIC_EGRESS_ADDR", envStringDefault("PROXY_RUNTIME_LOCAL_ADDR", ":1080")),
		LocalProtocol: normalizeConfigToken(envStringDefault("PROXY_RUNTIME_LOCAL_PROTOCOL", "http")),
		LocalUsername: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_LOCAL_USERNAME")),
		LocalPassword: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_LOCAL_PASSWORD")),
		SessionListener: SessionListenerConfig{
			AdvertisedHost: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_SESSION_ADVERTISED_HOST")),
		},
		ProviderHTTPProxy: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_PROVIDER_HTTP_PROXY")),
		ControlAuthToken:  strings.TrimSpace(os.Getenv("PROXY_RUNTIME_CONTROL_AUTH_TOKEN")),
		Provider:          normalizeConfigToken(envStringDefault("PROXY_RUNTIME_PROVIDER", ProviderTen24)),
		ProxyUsers:        proxyUsers,
		RefreshInterval:   refreshInterval,
		RequestTimeout:    requestTimeout,
		ProxyExitGeoURLs:  proxyExitGeoURLs("PROXY_RUNTIME_PROXY_EXIT_GEO_URLS"),
		EdgeCanaryTimeout: edgeCanaryTimeout,
		IPFraud: IPFraudConfig{
			Timeout:     ipFraudTimeout,
			CacheTTL:    ipFraudCacheTTL,
			KeyCooldown: ipFraudKeyCooldown,
		},
		Ten24: ten24.Config{
			APIURL:    strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_API_URL")),
			APIRegion: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_API_REGION")),
			APIFormat: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_API_FORMAT")),
			APITime:   strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_API_TIME")),
			APINum:    strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_API_NUM")),
			APIType:   strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_API_TYPE")),
			Username:  strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_USERNAME")),
			Password:  strings.TrimSpace(os.Getenv("PROXY_RUNTIME_1024_PASSWORD")),
			Protocol:  normalizeConfigToken(envStringDefault("PROXY_RUNTIME_1024_PROTOCOL", "http")),
		},
	}
	return cfg, cfg.validate()
}
