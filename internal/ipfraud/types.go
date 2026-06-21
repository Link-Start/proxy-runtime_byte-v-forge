package ipfraud

import (
	"context"
	"net/http"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/lookup"
)

type ProviderConfig struct {
	ID     string
	Kind   proxygatewayv1.ProxyIPFraudProviderKind
	Weight int
	Auth   AuthConfig
}

type AnonymousAuthConfig struct{}

type APIKeyAuthConfig struct {
	Keys      []string
	Placement string
	Name      string
}

type AuthConfig struct {
	Anonymous *AnonymousAuthConfig
	APIKey    *APIKeyAuthConfig
}

type Config struct {
	Providers   []ProviderConfig
	Timeout     time.Duration
	CacheTTL    time.Duration
	KeyCooldown time.Duration
	Clock       clock.Clock
}

type Plugin interface {
	lookup.PluginMeta[proxygatewayv1.ProxyIPFraudProviderKind]
	Auth(apiKeys []string, anonymous bool) AuthConfig
	New(client *http.Client, cfg ProviderConfig, cooldown time.Duration, clk clock.Clock) provider
}

type report struct {
	providerID     string
	providerName   string
	networkKind    proxygatewayv1.ProxyIPNetworkKind
	anonymizerKind proxygatewayv1.ProxyIPAnonymizerKind
	riskLevel      proxygatewayv1.ProxyIPFraudRiskLevel
	riskScore      float64
	signals        []proxygatewayv1.ProxyIPFraudSignal
	countryCode    string
	region         string
	city           string
	asn            string
	organization   string
	isp            string
}

type provider interface {
	lookup(ctx context.Context, ip string) (report, error)
}
