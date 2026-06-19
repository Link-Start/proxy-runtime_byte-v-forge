package accountproxy

import (
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type definitionPlugin struct{ definition Definition }

func NewDefinitionPlugin(definition Definition) Plugin {
	definition.ProviderID = normalizeProviderID(definition.ProviderID)
	definition.DisplayName = strings.TrimSpace(definition.DisplayName)
	definition.DefaultProtocol = strings.TrimSpace(definition.DefaultProtocol)
	return definitionPlugin{definition: definition}
}

func (p definitionPlugin) ID() string { return p.definition.ProviderID }

func (p definitionPlugin) DisplayName() string { return p.definition.DisplayName }

func (p definitionPlugin) Default() bool { return p.definition.Default }

func (p definitionPlugin) Descriptor(gateways []Gateway) *proxyruntimev1.ProxyProviderDescriptor {
	return descriptor(p.definition, gateways)
}

func (p definitionPlugin) GatewayProtocol(gateway Gateway) string {
	return GatewayProtocol(gateway, p.definition.DefaultProtocol)
}

func (p definitionPlugin) NewSessionProvider(cfg Config, client *http.Client, clk clock.Clock) (provider.SessionProvider, error) {
	cfg.ProviderID = p.definition.ProviderID
	if err := p.Validate(cfg); err != nil {
		return nil, err
	}
	if len(cfg.Gateways) == 0 {
		return nil, provider.ErrUnsupportedCapability
	}
	definition := p.definition
	definition.Gateways = cfg.Gateways
	return NewCredentialProvider(cfg, definition, clk), nil
}

func (p definitionPlugin) Validate(cfg Config) error {
	cfg.ProviderID = p.definition.ProviderID
	return validateConfig(cfg, p.definition)
}
