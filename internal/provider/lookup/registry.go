package lookup

import (
	"fmt"
	"sort"
)

// PluginMeta is the provider-plugin metadata contract shared by the IP lookup
// registries (ipfraud, ipgeo). Concrete plugins additionally expose
// package-specific Auth and New methods, so each package keeps its own Plugin
// interface that embeds this one.
type PluginMeta[K comparable] interface {
	Kind() K
	ProviderID() string
	DisplayName() string
	DefaultWeight() uint32
	SupportsAnonymous() bool
	SupportsAPIKey() bool
}

// Registry indexes provider plugins by their proto kind, ordered by descending
// default weight. K is the proto provider-kind enum whose zero value is the
// proto UNSPECIFIED sentinel; label names the domain for error messages.
type Registry[K comparable, P PluginMeta[K]] struct {
	label   string
	plugins map[K]P
	kinds   []K
}

func NewRegistry[K comparable, P PluginMeta[K]](label string, plugins ...P) (*Registry[K, P], error) {
	registry := &Registry[K, P]{label: label, plugins: make(map[K]P, len(plugins))}
	var zero K
	for _, plugin := range plugins {
		if any(plugin) == nil {
			return nil, fmt.Errorf("%s provider plugin is required", label)
		}
		kind := plugin.Kind()
		if kind == zero {
			return nil, fmt.Errorf("%s provider plugin kind is required", label)
		}
		if _, exists := registry.plugins[kind]; exists {
			return nil, fmt.Errorf("duplicate %s provider kind %v", label, kind)
		}
		registry.plugins[kind] = plugin
		registry.kinds = append(registry.kinds, kind)
	}
	registry.sortKinds()
	return registry, nil
}

func (r *Registry[K, P]) PluginForKind(kind K) (P, bool) {
	if r == nil {
		var zero P
		return zero, false
	}
	plugin, ok := r.plugins[kind]
	return plugin, ok
}

func (r *Registry[K, P]) IsProviderKindSupported(kind K) bool {
	_, ok := r.PluginForKind(kind)
	return ok
}

func (r *Registry[K, P]) DefaultProviderID(kind K) string {
	plugin, ok := r.PluginForKind(kind)
	if !ok {
		return ""
	}
	return plugin.ProviderID()
}

// Kinds returns the registered provider kinds ordered by descending default
// weight.
func (r *Registry[K, P]) Kinds() []K {
	if r == nil {
		return nil
	}
	return r.kinds
}

func (r *Registry[K, P]) sortKinds() {
	sort.Slice(r.kinds, func(i, j int) bool {
		return r.plugins[r.kinds[i]].DefaultWeight() > r.plugins[r.kinds[j]].DefaultWeight()
	})
}
