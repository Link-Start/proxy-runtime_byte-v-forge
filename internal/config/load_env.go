package config

import (
	"os"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/ten24"
)

func LoadFromEnv() (Config, error) {
	mihomoConfigDir := envx.StringDefault("PROXY_RUNTIME_MIHOMO_CONFIG_DIR", "/var/lib/byte-v-forge/proxy-runtime/mihomo")
	cfg := Config{
		RuntimeAddr:   envx.StringDefault("PROXY_RUNTIME_ADDR", ":8080"),
		DataDir:       envx.StringDefault("PROXY_RUNTIME_DATA_DIR", "/var/lib/byte-v-forge/proxy-runtime"),
		PostgresDSN:   firstNonEmpty(strings.TrimSpace(os.Getenv("PROXY_RUNTIME_POSTGRES_DSN")), strings.TrimSpace(os.Getenv("PG_DSN"))),
		RedisURL:      firstNonEmpty(strings.TrimSpace(os.Getenv("PROXY_RUNTIME_REDIS_URL")), strings.TrimSpace(os.Getenv("PLATFORM_REDIS_URL"))),
		EncryptionKey: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_ENCRYPTION_KEY")),
		Mihomo: MihomoConfig{
			Path:                envx.StringDefault("PROXY_RUNTIME_MIHOMO_PATH", "mihomo"),
			ConfigDir:           mihomoConfigDir,
			APIAddr:             envx.StringDefault("PROXY_RUNTIME_MIHOMO_API_ADDR", "127.0.0.1:18901"),
			DashboardDir:        envx.StringDefault("PROXY_RUNTIME_MIHOMO_DASHBOARD_DIR", "/app/dashboard/metacubexd"),
			DashboardURL:        strings.TrimSpace(os.Getenv("PROXY_RUNTIME_MIHOMO_DASHBOARD_URL")),
			HealthCheckURL:      envx.StringDefault("PROXY_RUNTIME_MIHOMO_HEALTH_CHECK_URL", "https://www.gstatic.com/generate_204"),
			HealthCheckInterval: envx.DurationSeconds("PROXY_RUNTIME_MIHOMO_HEALTH_CHECK_INTERVAL_SECONDS", 300*time.Second),
			HealthCheckTimeout:  envx.DurationSeconds("PROXY_RUNTIME_MIHOMO_HEALTH_CHECK_TIMEOUT_SECONDS", 5*time.Second),
		},
		CommonEgressAddr: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_COMMON_EGRESS_ADDR")),
		LocalAddr:        envx.StringDefault("PROXY_RUNTIME_DYNAMIC_EGRESS_ADDR", envx.StringDefault("PROXY_RUNTIME_LOCAL_ADDR", ":1080")),
		LocalProtocol:    normalizeConfigToken(envx.StringDefault("PROXY_RUNTIME_LOCAL_PROTOCOL", "http")),
		LocalUsername:    strings.TrimSpace(os.Getenv("PROXY_RUNTIME_LOCAL_USERNAME")),
		LocalPassword:    strings.TrimSpace(os.Getenv("PROXY_RUNTIME_LOCAL_PASSWORD")),
		SessionListener: SessionListenerConfig{
			AdvertisedHost: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_SESSION_ADVERTISED_HOST")),
		},
		ProviderHTTPProxy: strings.TrimSpace(os.Getenv("PROXY_RUNTIME_PROVIDER_HTTP_PROXY")),
		Provider:          normalizeConfigToken(envx.StringDefault("PROXY_RUNTIME_PROVIDER", ProviderTen24)),
		ProxyUsers:        envProxyUsers("PROXY_RUNTIME_PROXY_USERS_JSON"),
		RefreshInterval:   envx.DurationSeconds("PROXY_RUNTIME_REFRESH_SECONDS", 300*time.Second),
		RequestTimeout:    envx.DurationSeconds("PROXY_RUNTIME_REQUEST_TIMEOUT_SECONDS", 10*time.Second),
		ProxyExitGeoURLs:  proxyExitGeoURLs("PROXY_RUNTIME_PROXY_EXIT_GEO_URLS"),
		EdgeCanaryTimeout: envx.DurationSeconds("PROXY_RUNTIME_EDGE_CANARY_TIMEOUT_SECONDS", 10*time.Second),
		IPFraud: IPFraudConfig{
			Timeout:     envx.DurationSeconds("PROXY_RUNTIME_IP_FRAUD_TIMEOUT_SECONDS", 10*time.Second),
			CacheTTL:    envx.DurationSeconds("PROXY_RUNTIME_IP_FRAUD_CACHE_TTL_SECONDS", 10*time.Minute),
			KeyCooldown: envx.DurationSeconds("PROXY_RUNTIME_IP_FRAUD_KEY_COOLDOWN_SECONDS", 24*time.Hour),
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
			Protocol:  normalizeConfigToken(envx.StringDefault("PROXY_RUNTIME_1024_PROTOCOL", "http")),
		},
	}
	return cfg, cfg.validate()
}
