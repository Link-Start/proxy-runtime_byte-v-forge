package ipgeo

import (
	"context"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ProviderConfig struct {
	ID     string
	Kind   proxyruntimev1.ProxyIPGeoProviderKind
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
	Kind() proxyruntimev1.ProxyIPGeoProviderKind
	ProviderID() string
	DisplayName() string
	DefaultWeight() uint32
	SupportsAnonymous() bool
	SupportsAPIKey() bool
	Auth(apiKeys []string, anonymous bool) AuthConfig
	New(client *http.Client, cfg ProviderConfig) provider
}

type provider interface {
	Lookup(ctx context.Context, ip string) (*proxyruntimev1.ProxyExitGeo, error)
}
