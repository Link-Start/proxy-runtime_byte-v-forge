package accountproxy

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type Registry struct {
	plugins map[string]Plugin
	ids     []string
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	registry := &Registry{plugins: make(map[string]Plugin, len(plugins))}
	for _, plugin := range plugins {
		if plugin == nil {
			return nil, fmt.Errorf("proxy account plugin is required")
		}
		id := normalizeProviderID(plugin.ID())
		if id == "" {
			return nil, fmt.Errorf("proxy account plugin id is required")
		}
		if _, exists := registry.plugins[id]; exists {
			return nil, fmt.Errorf("duplicate proxy account plugin %q", id)
		}
		registry.plugins[id] = plugin
		registry.ids = append(registry.ids, id)
	}
	sort.Strings(registry.ids)
	return registry, nil
}

func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(Ten24Plugin(), B2ProxyPlugin(), CliproxyPlugin())
}

func (r *Registry) Get(providerID string) (Plugin, bool) {
	if r == nil {
		return nil, false
	}
	plugin, ok := r.plugins[normalizeProviderID(providerID)]
	return plugin, ok
}

func (r *Registry) IsSupported(providerID string) bool {
	_, ok := r.Get(providerID)
	return ok
}

func (r *Registry) SupportsRuntimeGeoTargeting(providerID string) bool {
	plugin, ok := r.Get(providerID)
	return ok && plugin.SupportsRuntimeGeoTargeting()
}

func (r *Registry) Descriptors(gateways map[string][]Gateway) []*proxyruntimev1.ProxyProviderDescriptor {
	if r == nil {
		return nil
	}
	out := make([]*proxyruntimev1.ProxyProviderDescriptor, 0, len(r.ids))
	for _, id := range r.ids {
		out = append(out, r.plugins[id].Descriptor(gateways[id]))
	}
	return out
}

func (r *Registry) NewProvider(cfg Config, client *http.Client) (provider.Provider, error) {
	plugin, ok := r.Get(cfg.ProviderID)
	if !ok {
		return nil, fmt.Errorf("unsupported provider_id %q", cfg.ProviderID)
	}
	return plugin.NewProvider(cfg, client)
}

func (r *Registry) Validate(cfg Config) error {
	plugin, ok := r.Get(cfg.ProviderID)
	if !ok {
		return fmt.Errorf("unsupported provider_id %q", cfg.ProviderID)
	}
	return plugin.Validate(cfg)
}

func (r *Registry) GatewayProtocolForProvider(providerID string, gateway Gateway) (string, bool) {
	plugin, ok := r.Get(providerID)
	if !ok {
		return "", false
	}
	return plugin.GatewayProtocol(gateway), true
}

func (r *Registry) DynamicSource(providerID string, displayName string, accountID string, gateways []Gateway) (*proxyruntimev1.ProxySourceDescriptor, error) {
	plugin, ok := r.Get(providerID)
	if !ok {
		return nil, fmt.Errorf("unsupported provider_id %q", providerID)
	}
	return plugin.DynamicSource(accountID, displayName, gateways), nil
}

func normalizeProviderID(value string) string { return strings.TrimSpace(value) }
