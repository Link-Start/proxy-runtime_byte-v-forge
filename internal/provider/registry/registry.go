package registry

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/ten24"
)

type BuildContext struct {
	Config     config.Config
	HTTPClient *http.Client
}

type Plugin interface {
	ID() string
	NewPoolProvider(BuildContext) (provider.PoolProvider, error)
}

type plugin struct {
	id      string
	provide func(BuildContext) (provider.PoolProvider, error)
}

type Registry struct {
	plugins        map[string]Plugin
	ids            []string
	accountPlugins map[string]accountproxy.Plugin
	accountIDs     []string
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	registry := &Registry{plugins: make(map[string]Plugin, len(plugins)), accountPlugins: map[string]accountproxy.Plugin{}}
	for _, plugin := range plugins {
		if plugin == nil {
			return nil, fmt.Errorf("proxy provider plugin is required")
		}
		id := normalizeProviderID(plugin.ID())
		if id == "" {
			return nil, fmt.Errorf("proxy provider plugin id is required")
		}
		if _, exists := registry.plugins[id]; exists {
			return nil, fmt.Errorf("duplicate proxy provider plugin %q", id)
		}
		registry.plugins[id] = plugin
		registry.ids = append(registry.ids, id)
	}
	sort.Strings(registry.ids)
	return registry, nil
}

func NewDefaultRegistry() (*Registry, error) {
	registry, err := NewRegistry(
		NewPoolPlugin(config.ProviderNone, func(BuildContext) (provider.PoolProvider, error) {
			return provider.Empty{}, nil
		}),
		NewPoolPlugin(config.ProviderTen24, func(ctx BuildContext) (provider.PoolProvider, error) {
			return ten24.New(ctx.Config.Ten24, ctx.HTTPClient), nil
		}),
	)
	if err != nil {
		return nil, err
	}
	if err := registry.registerAccountPlugins(accountproxy.Ten24Plugin(), accountproxy.B2ProxyPlugin(), accountproxy.CliproxyPlugin()); err != nil {
		return nil, err
	}
	return registry, nil
}

func NewPoolPlugin(id string, provide func(BuildContext) (provider.PoolProvider, error)) Plugin {
	return plugin{id: normalizeProviderID(id), provide: provide}
}

func (p plugin) ID() string { return p.id }

func (p plugin) NewPoolProvider(ctx BuildContext) (provider.PoolProvider, error) {
	if p.provide == nil {
		return nil, fmt.Errorf("proxy provider plugin %q has no provider factory", p.id)
	}
	return p.provide(ctx)
}

func (r *Registry) NewPoolProvider(cfg config.Config, client *http.Client) (provider.PoolProvider, error) {
	if r == nil {
		return nil, config.ErrUnsupportedProvider
	}
	plugin, ok := r.plugins[normalizeProviderID(cfg.Provider)]
	if !ok {
		return nil, config.ErrUnsupportedProvider
	}
	return plugin.NewPoolProvider(BuildContext{Config: cfg, HTTPClient: client})
}

func (r *Registry) registerAccountPlugins(plugins ...accountproxy.Plugin) error {
	if r == nil {
		return fmt.Errorf("proxy provider registry is required")
	}
	for _, plugin := range plugins {
		if plugin == nil {
			return fmt.Errorf("proxy account plugin is required")
		}
		id := normalizeProviderID(plugin.ID())
		if id == "" {
			return fmt.Errorf("proxy account plugin id is required")
		}
		if _, exists := r.accountPlugins[id]; exists {
			return fmt.Errorf("duplicate proxy account plugin %q", id)
		}
		r.accountPlugins[id] = plugin
		r.accountIDs = append(r.accountIDs, id)
	}
	sort.Strings(r.accountIDs)
	return nil
}

func (r *Registry) accountPlugin(providerID string) (accountproxy.Plugin, bool) {
	if r == nil {
		return nil, false
	}
	plugin, ok := r.accountPlugins[normalizeProviderID(providerID)]
	return plugin, ok
}

func (r *Registry) IsSupported(providerID string) bool {
	_, ok := r.accountPlugin(providerID)
	return ok
}

func (r *Registry) Descriptors(gateways map[string][]accountproxy.Gateway) []*proxyruntimev1.ProxyProviderDescriptor {
	if r == nil {
		return nil
	}
	out := make([]*proxyruntimev1.ProxyProviderDescriptor, 0, len(r.accountIDs))
	for _, id := range r.accountIDs {
		out = append(out, r.accountPlugins[id].Descriptor(gateways[id]))
	}
	return out
}

func (r *Registry) NewSessionProvider(cfg accountproxy.Config, client *http.Client) (provider.SessionProvider, error) {
	plugin, ok := r.accountPlugin(cfg.ProviderID)
	if !ok {
		return nil, fmt.Errorf("unsupported provider_id %q", cfg.ProviderID)
	}
	return plugin.NewSessionProvider(cfg, client)
}

func (r *Registry) Validate(cfg accountproxy.Config) error {
	plugin, ok := r.accountPlugin(cfg.ProviderID)
	if !ok {
		return fmt.Errorf("unsupported provider_id %q", cfg.ProviderID)
	}
	return plugin.Validate(cfg)
}

func (r *Registry) GatewayProtocolForProvider(providerID string, gateway accountproxy.Gateway) (string, bool) {
	plugin, ok := r.accountPlugin(providerID)
	if !ok {
		return "", false
	}
	return plugin.GatewayProtocol(gateway), true
}

func normalizeProviderID(value string) string { return strings.TrimSpace(value) }
