package ipgeo

import (
	"context"
	"net/http"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/lookup"
)

type ProviderConfig struct {
	ID     string
	Kind   proxygatewayv1.ProxyIPGeoProviderKind
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

type Plugin interface {
	lookup.PluginMeta[proxygatewayv1.ProxyIPGeoProviderKind]
	Auth(apiKeys []string, anonymous bool) AuthConfig
	New(client *http.Client, cfg ProviderConfig) provider
}

type provider interface {
	Lookup(ctx context.Context, ip string) (*proxygatewayv1.ProxyExitGeo, error)
}
