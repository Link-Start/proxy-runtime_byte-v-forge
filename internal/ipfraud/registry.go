package ipfraud

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/lookup"
)

type Registry struct {
	*lookup.Registry[proxygatewayv1.ProxyIPFraudProviderKind, Plugin]
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	base, err := lookup.NewRegistry[proxygatewayv1.ProxyIPFraudProviderKind, Plugin]("IP fraud", plugins...)
	if err != nil {
		return nil, err
	}
	return &Registry{base}, nil
}

func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(ipQualityScorePlugin{}, ipapiPlugin{}, abuseIPDBPlugin{})
}

func (r *Registry) ProviderDescriptors() []*proxygatewayv1.ProxyIPFraudProviderDescriptor {
	if r == nil {
		return nil
	}
	kinds := r.Kinds()
	out := make([]*proxygatewayv1.ProxyIPFraudProviderDescriptor, 0, len(kinds))
	for _, kind := range kinds {
		plugin, ok := r.PluginForKind(kind)
		if !ok {
			continue
		}
		out = append(out, &proxygatewayv1.ProxyIPFraudProviderDescriptor{ProviderId: plugin.ProviderID(), DisplayName: plugin.DisplayName(), DefaultWeight: plugin.DefaultWeight(), Kind: kind, SupportsAnonymous: plugin.SupportsAnonymous(), SupportsApiKey: plugin.SupportsAPIKey()})
	}
	return out
}
