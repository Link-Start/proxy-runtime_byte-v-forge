package accountproxy

import (
	"net/http"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

const (
	ProviderTen24    = "1024proxy"
	ProviderB2Proxy  = "b2proxy"
	ProviderCliproxy = "cliproxy"

	defaultStickyMinutes = 10
	minStickyMinutes     = 1
	maxStickyMinutes     = 120
)

type Config struct {
	ProviderID string
	Username   string
	Password   string
	Gateways   []Gateway
}

type Gateway struct {
	ID          string
	EndpointURL string
}

type Plugin interface {
	ID() string
	DisplayName() string
	Default() bool
	Descriptor(gateways []Gateway) *proxygatewayv1.ProxyProviderDescriptor
	GatewayProtocol(gateway Gateway) string
	NewSessionProvider(cfg Config, client *http.Client, clk clock.Clock) (provider.SessionProvider, error)
	Validate(cfg Config) error
}

type UsernameBuilder func(base string, policy *proxygatewayv1.ProxySessionPolicy, sessionID string) string
type SessionIDGenerator func() (string, error)

type Definition struct {
	ProviderID               string
	DisplayName              string
	Default                  bool
	DefaultProtocol          string
	Protocols                []string
	Gateways                 []Gateway
	UsernameParameterSession bool
	BuildUsername            UsernameBuilder
	GenerateSessionID        SessionIDGenerator
}

func normalizeProviderID(value string) string { return strings.TrimSpace(value) }
