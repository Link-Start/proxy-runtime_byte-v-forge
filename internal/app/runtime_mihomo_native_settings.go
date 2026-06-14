package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func mihomoNativeSettings(ctx context.Context, runtime *Runtime) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if runtime == nil || runtime.settings == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	return runtime.settings.loadMihomoNative(ctx)
}

func mihomoNativeSettingsEmpty(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) bool {
	return view == nil || len(view.GetFixedProxies()) == 0 && len(view.GetSubscriptions()) == 0
}

func mihomoNativeSettingsFromConfig(config mihomoNativeConfigFile) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	view := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	if len(config.FixedProxies) > 0 {
		for _, proxy := range config.FixedProxies {
			item := normalizeMihomoNativeFixedProxy(proxy, nil)
			if item.Name == "" || item.URI == "" {
				continue
			}
			view.FixedProxies = append(view.FixedProxies, protoMihomoNativeFixedProxy(item))
		}
	} else {
		for _, proxy := range config.Proxies {
			name := jsonStringValue(proxy["name"])
			proxyType := jsonStringValue(proxy["type"])
			if name == "" {
				continue
			}
			uri := mihomoNativeProxyURI(proxy)
			if uri == "" {
				continue
			}
			view.FixedProxies = append(view.FixedProxies, protoMihomoNativeFixedProxy(mihomoNativeFixedProxy{ID: nativeStableID("fixed", uri), Name: name, Type: proxyType, URI: uri}))
		}
	}
	if len(config.Subscriptions) > 0 {
		for _, subscription := range config.Subscriptions {
			item := normalizeMihomoNativeSubscription(subscription, nil)
			if item.Name == "" || item.URL == "" {
				continue
			}
			view.Subscriptions = append(view.Subscriptions, protoMihomoNativeSubscription(item))
		}
	} else {
		for name, provider := range config.ProxyProviders {
			if strings.TrimSpace(provider.URL) == "" {
				continue
			}
			view.Subscriptions = append(view.Subscriptions, protoMihomoNativeSubscription(mihomoNativeSubscription{ID: nativeStableID("sub", provider.URL), Name: name, URL: provider.URL}))
		}
	}
	return normalizeMihomoNativeSettings(view)
}

func mihomoNativeConfigFileFromSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (mihomoNativeConfigFile, error) {
	view = normalizeMihomoNativeSettings(view)
	config := mihomoNativeConfigFile{
		FixedProxies:   make([]mihomoNativeFixedProxy, 0, len(view.GetFixedProxies())),
		Proxies:        make([]map[string]any, 0, len(view.GetFixedProxies())),
		Subscriptions:  make([]mihomoNativeSubscription, 0, len(view.GetSubscriptions())),
		ProxyProviders: map[string]mihomoNativeProvider{},
	}
	for _, item := range view.GetFixedProxies() {
		proxy := nativeFixedProxyFromProto(item)
		rendered, err := mihomoNativeProxyFromURI(proxy.Name, proxy.URI)
		if err != nil {
			return mihomoNativeConfigFile{}, err
		}
		proxy.Type = jsonStringValue(rendered["type"])
		config.FixedProxies = append(config.FixedProxies, proxy)
		config.Proxies = append(config.Proxies, rendered)
	}
	for _, item := range view.GetSubscriptions() {
		provider, subscription, err := mihomoNativeSubscriptionProvider(nativeSubscriptionFromProto(item))
		if err != nil {
			return mihomoNativeConfigFile{}, err
		}
		config.Subscriptions = append(config.Subscriptions, subscription)
		config.ProxyProviders[subscription.Name] = provider
	}
	return config, nil
}

func protoMihomoNativeFixedProxy(item mihomoNativeFixedProxy) *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy{Id: item.ID, Name: item.Name, Type: item.Type, Uri: item.URI}
}

func protoMihomoNativeSubscription(item mihomoNativeSubscription) *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeSubscription{Id: item.ID, Name: item.Name, Url: item.URL}
}

func normalizeMihomoNativeSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	out := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{
		FixedProxies:  make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(view.GetFixedProxies())),
		Subscriptions: make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(view.GetSubscriptions())),
	}
	seenFixed := map[string]struct{}{}
	for _, item := range view.GetFixedProxies() {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(item), nil)
		key := firstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seenFixed[key]; exists {
			continue
		}
		seenFixed[key] = struct{}{}
		out.FixedProxies = append(out.FixedProxies, protoMihomoNativeFixedProxy(normalized))
	}
	seenSubscriptions := map[string]struct{}{}
	for _, item := range view.GetSubscriptions() {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(item), nil)
		key := firstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seenSubscriptions[key]; exists {
			continue
		}
		seenSubscriptions[key] = struct{}{}
		out.Subscriptions = append(out.Subscriptions, protoMihomoNativeSubscription(normalized))
	}
	return out
}

func nativeFixedProxyFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) mihomoNativeFixedProxy {
	if item == nil {
		return mihomoNativeFixedProxy{}
	}
	return mihomoNativeFixedProxy{ID: item.GetId(), Name: item.GetName(), Type: item.GetType(), URI: item.GetUri()}
}

func nativeSubscriptionFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) mihomoNativeSubscription {
	if item == nil {
		return mihomoNativeSubscription{}
	}
	return mihomoNativeSubscription{ID: item.GetId(), Name: item.GetName(), URL: item.GetUrl()}
}

func preserveNativeGroups(groups []mihomoNativeGroup) []mihomoNativeGroup {
	out := make([]mihomoNativeGroup, 0, len(groups))
	for _, group := range groups {
		if strings.TrimSpace(group.Name) == "" || group.Name == mihomoFixedProxyGroupName {
			continue
		}
		out = append(out, group)
	}
	return out
}

func currentFixedProxyIndexes(config mihomoNativeConfigFile) (map[string]mihomoNativeFixedProxy, map[string]mihomoNativeFixedProxy) {
	byID := map[string]mihomoNativeFixedProxy{}
	byName := map[string]mihomoNativeFixedProxy{}
	for _, item := range config.FixedProxies {
		normalized := normalizeMihomoNativeFixedProxy(item, nil)
		if normalized.ID != "" {
			byID[normalized.ID] = normalized
		}
		if normalized.Name != "" {
			byName[normalized.Name] = normalized
		}
	}
	for _, proxy := range config.Proxies {
		name := jsonStringValue(proxy["name"])
		uri := mihomoNativeProxyURI(proxy)
		if name == "" || uri == "" {
			continue
		}
		normalized := mihomoNativeFixedProxy{ID: nativeStableID("fixed", uri), Name: name, Type: jsonStringValue(proxy["type"]), URI: uri}
		if _, exists := byID[normalized.ID]; !exists {
			byID[normalized.ID] = normalized
		}
		if _, exists := byName[normalized.Name]; !exists {
			byName[normalized.Name] = normalized
		}
	}
	return byID, byName
}

func currentSubscriptionIndexes(config mihomoNativeConfigFile) (map[string]mihomoNativeSubscription, map[string]mihomoNativeSubscription) {
	byID := map[string]mihomoNativeSubscription{}
	byName := map[string]mihomoNativeSubscription{}
	for _, item := range config.Subscriptions {
		normalized := normalizeMihomoNativeSubscription(item, nil)
		if normalized.ID != "" {
			byID[normalized.ID] = normalized
		}
		if normalized.Name != "" {
			byName[normalized.Name] = normalized
		}
	}
	for name, provider := range config.ProxyProviders {
		url := strings.TrimSpace(provider.URL)
		if strings.TrimSpace(name) == "" || url == "" {
			continue
		}
		normalized := mihomoNativeSubscription{ID: nativeStableID("sub", url), Name: strings.TrimSpace(name), URL: url}
		if _, exists := byID[normalized.ID]; !exists {
			byID[normalized.ID] = normalized
		}
		if _, exists := byName[normalized.Name]; !exists {
			byName[normalized.Name] = normalized
		}
	}
	return byID, byName
}

func normalizeMihomoNativeFixedProxy(item mihomoNativeFixedProxy, currentByName map[string]mihomoNativeFixedProxy) mihomoNativeFixedProxy {
	item.Name = strings.TrimSpace(item.Name)
	item.URI = strings.TrimSpace(item.URI)
	item.Type = strings.TrimSpace(item.Type)
	item.ID = runtimeSafeID(item.ID)
	if item.ID == "" && currentByName != nil {
		item.ID = runtimeSafeID(currentByName[item.Name].ID)
	}
	if item.ID == "" {
		item.ID = nativeStableID("fixed", item.URI)
	}
	return item
}

func normalizeMihomoNativeSubscription(item mihomoNativeSubscription, currentByName map[string]mihomoNativeSubscription) mihomoNativeSubscription {
	item.Name = strings.TrimSpace(item.Name)
	item.URL = strings.TrimSpace(item.URL)
	item.ID = runtimeSafeID(item.ID)
	if item.ID == "" && currentByName != nil {
		item.ID = runtimeSafeID(currentByName[item.Name].ID)
	}
	if item.ID == "" {
		item.ID = nativeStableID("sub", item.URL)
	}
	return item
}

func nativeStableID(prefix string, source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(source))
	return runtimeSafeID(prefix + "-" + hex.EncodeToString(sum[:])[:12])
}

func mihomoNativeSubscriptionProvider(item mihomoNativeSubscription) (mihomoNativeProvider, mihomoNativeSubscription, error) {
	id := runtimeSafeID(item.ID)
	name := strings.TrimSpace(item.Name)
	rawURL := strings.TrimSpace(item.URL)
	if id == "" {
		id = nativeStableID("sub", rawURL)
	}
	if name == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument("subscription name is required", nil)
	}
	if rawURL == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument(fmt.Sprintf("subscription %q url is required", name), nil)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument(fmt.Sprintf("subscription %q url is invalid", name), nil)
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument(fmt.Sprintf("subscription %q url must use http or https", name), nil)
	}
	return mihomoNativeProvider{
		Type:     "http",
		URL:      rawURL,
		Path:     filepath.Join("providers", id+".yaml"),
		Interval: defaultProviderHealthPeriod,
		Header:   map[string][]string{"User-Agent": {defaultProviderUserAgent}},
		HealthCheck: &mihomoNativeHealthCheck{
			Enable:         true,
			URL:            defaultProviderHealthURL,
			Interval:       defaultProviderHealthPeriod,
			Timeout:        defaultProviderHealthWait,
			Lazy:           true,
			ExpectedStatus: 204,
		},
	}, mihomoNativeSubscription{ID: id, Name: name, URL: rawURL}, nil
}
