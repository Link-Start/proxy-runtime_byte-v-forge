package accountproxy

import (
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
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
	Descriptor(gateways []Gateway) *proxyruntimev1.ProxyProviderDescriptor
	GatewayProtocol(gateway Gateway) string
	NewSessionProvider(cfg Config, client *http.Client) (provider.SessionProvider, error)
	Validate(cfg Config) error
}

type UsernameBuilder func(base string, policy *proxyruntimev1.ProxySessionPolicy, sessionID string) string
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
