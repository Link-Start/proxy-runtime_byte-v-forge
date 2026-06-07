package ipgeo

import (
	"fmt"
	"sort"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type Registry struct {
	plugins map[proxyruntimev1.ProxyIPGeoProviderKind]Plugin
	kinds   []proxyruntimev1.ProxyIPGeoProviderKind
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	registry := &Registry{plugins: make(map[proxyruntimev1.ProxyIPGeoProviderKind]Plugin, len(plugins))}
	for _, plugin := range plugins {
		if plugin == nil {
			return nil, fmt.Errorf("IP geo provider plugin is required")
		}
		kind := plugin.Kind()
		if kind == proxyruntimev1.ProxyIPGeoProviderKind_PROXY_IP_GEO_PROVIDER_KIND_UNSPECIFIED {
			return nil, fmt.Errorf("IP geo provider plugin kind is required")
		}
		if _, exists := registry.plugins[kind]; exists {
			return nil, fmt.Errorf("duplicate IP geo provider kind %s", kind.String())
		}
		registry.plugins[kind] = plugin
		registry.kinds = append(registry.kinds, kind)
	}
	registry.sortKinds()
	return registry, nil
}

func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(ipinfoPlugin{}, ip2LocationPlugin{}, ipAPIComPlugin{})
}

func (r *Registry) PluginForKind(kind proxyruntimev1.ProxyIPGeoProviderKind) (Plugin, bool) {
	if r == nil {
		return nil, false
	}
	plugin, ok := r.plugins[kind]
	return plugin, ok
}

func (r *Registry) IsProviderKindSupported(kind proxyruntimev1.ProxyIPGeoProviderKind) bool {
	_, ok := r.PluginForKind(kind)
	return ok
}

func (r *Registry) DefaultProviderID(kind proxyruntimev1.ProxyIPGeoProviderKind) string {
	plugin, ok := r.PluginForKind(kind)
	if !ok {
		return ""
	}
	return plugin.ProviderID()
}

func (r *Registry) ProviderDescriptors() []*proxyruntimev1.ProxyIPGeoProviderDescriptor {
	if r == nil {
		return nil
	}
	out := make([]*proxyruntimev1.ProxyIPGeoProviderDescriptor, 0, len(r.kinds))
	for _, kind := range r.kinds {
		plugin := r.plugins[kind]
		out = append(out, &proxyruntimev1.ProxyIPGeoProviderDescriptor{ProviderId: plugin.ProviderID(), DisplayName: plugin.DisplayName(), DefaultWeight: plugin.DefaultWeight(), Kind: kind, SupportsAnonymous: plugin.SupportsAnonymous(), SupportsApiKey: plugin.SupportsAPIKey()})
	}
	return out
}

func (r *Registry) sortKinds() {
	sort.Slice(r.kinds, func(i, j int) bool {
		return r.plugins[r.kinds[i]].DefaultWeight() > r.plugins[r.kinds[j]].DefaultWeight()
	})
}
