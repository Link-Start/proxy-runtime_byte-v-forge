package registry

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/ten24"
)

type BuildContext struct {
	Config     config.Config
	HTTPClient *http.Client
}

type Plugin interface {
	ID() string
	NewProvider(BuildContext) (provider.Provider, error)
}

type plugin struct {
	id      string
	provide func(BuildContext) (provider.Provider, error)
}

type Registry struct {
	plugins map[string]Plugin
	ids     []string
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	registry := &Registry{plugins: make(map[string]Plugin, len(plugins))}
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
	return NewRegistry(
		NewPlugin(config.ProviderNone, func(BuildContext) (provider.Provider, error) {
			return provider.Empty{}, nil
		}),
		NewPlugin(config.ProviderStatic, func(ctx BuildContext) (provider.Provider, error) {
			return provider.NewStatic(ctx.Config.SimpleProxies)
		}),
		NewPlugin(config.ProviderTen24, func(ctx BuildContext) (provider.Provider, error) {
			return ten24.New(ctx.Config.Ten24, ctx.HTTPClient), nil
		}),
	)
}

func NewPlugin(id string, provide func(BuildContext) (provider.Provider, error)) Plugin {
	return plugin{id: normalizeProviderID(id), provide: provide}
}

func (p plugin) ID() string { return p.id }

func (p plugin) NewProvider(ctx BuildContext) (provider.Provider, error) {
	if p.provide == nil {
		return nil, fmt.Errorf("proxy provider plugin %q has no provider factory", p.id)
	}
	return p.provide(ctx)
}

func (r *Registry) NewProvider(cfg config.Config, client *http.Client) (provider.Provider, error) {
	if r == nil {
		return nil, config.ErrUnsupportedProvider
	}
	plugin, ok := r.plugins[normalizeProviderID(cfg.Provider)]
	if !ok {
		return nil, config.ErrUnsupportedProvider
	}
	return plugin.NewProvider(BuildContext{Config: cfg, HTTPClient: client})
}

func normalizeProviderID(value string) string { return strings.TrimSpace(value) }
