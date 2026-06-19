package ipgeo

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/lookup"
)

type Registry struct {
	*lookup.Registry[proxyruntimev1.ProxyIPGeoProviderKind, Plugin]
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	base, err := lookup.NewRegistry[proxyruntimev1.ProxyIPGeoProviderKind, Plugin]("IP geo", plugins...)
	if err != nil {
		return nil, err
	}
	return &Registry{base}, nil
}

func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(ipinfoPlugin{}, ip2LocationPlugin{}, ipAPIComPlugin{})
}

func (r *Registry) ProviderDescriptors() []*proxyruntimev1.ProxyIPGeoProviderDescriptor {
	if r == nil {
		return nil
	}
	kinds := r.Kinds()
	out := make([]*proxyruntimev1.ProxyIPGeoProviderDescriptor, 0, len(kinds))
	for _, kind := range kinds {
		plugin, ok := r.PluginForKind(kind)
		if !ok {
			continue
		}
		out = append(out, &proxyruntimev1.ProxyIPGeoProviderDescriptor{ProviderId: plugin.ProviderID(), DisplayName: plugin.DisplayName(), DefaultWeight: plugin.DefaultWeight(), Kind: kind, SupportsAnonymous: plugin.SupportsAnonymous(), SupportsApiKey: plugin.SupportsAPIKey()})
	}
	return out
}
