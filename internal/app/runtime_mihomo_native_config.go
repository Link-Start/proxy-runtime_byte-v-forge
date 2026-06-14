package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	mihomoNativeFileName        = "native.json"
	mihomoFixedProxyGroupName   = "固定代理"
	defaultProviderHealthURL    = "https://www.gstatic.com/generate_204"
	defaultProviderHealthPeriod = 300
	defaultProviderHealthWait   = 5000
	defaultProviderUserAgent    = "mihomo/1.18.3"
)

type mihomoNativeFixedProxy struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	URI  string `json:"uri"`
}

type mihomoNativeSubscription struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type mihomoNativeConfigFile struct {
	FixedProxies   []mihomoNativeFixedProxy        `json:"fixed_proxies,omitempty"`
	Proxies        []map[string]any                `json:"proxies,omitempty"`
	Subscriptions  []mihomoNativeSubscription      `json:"subscriptions,omitempty"`
	ProxyProviders map[string]mihomoNativeProvider `json:"proxy-providers,omitempty"`
	ProxyGroups    []mihomoNativeGroup             `json:"proxy-groups,omitempty"`
	Rules          []string                        `json:"rules,omitempty"`
}

type mihomoNativeProvider struct {
	Type        string                   `json:"type"`
	URL         string                   `json:"url,omitempty"`
	Path        string                   `json:"path,omitempty"`
	Interval    int                      `json:"interval,omitempty"`
	Filter      string                   `json:"filter,omitempty"`
	Exclude     string                   `json:"exclude-filter,omitempty"`
	HealthCheck *mihomoNativeHealthCheck `json:"health-check,omitempty"`
	Header      map[string][]string      `json:"header,omitempty"`
	Override    map[string]any           `json:"override,omitempty"`
}

type mihomoNativeHealthCheck struct {
	Enable         bool   `json:"enable"`
	URL            string `json:"url,omitempty"`
	Interval       int    `json:"interval,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	Lazy           bool   `json:"lazy"`
	ExpectedStatus uint32 `json:"expected-status,omitempty"`
}

type mihomoNativeGroup struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Proxies        []string `json:"proxies,omitempty"`
	Use            []string `json:"use,omitempty"`
	Filter         string   `json:"filter,omitempty"`
	URL            string   `json:"url,omitempty"`
	Interval       int      `json:"interval,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	Strategy       string   `json:"strategy,omitempty"`
	Lazy           bool     `json:"lazy"`
	ExpectedStatus uint32   `json:"expected-status,omitempty"`
	Hidden         bool     `json:"hidden,omitempty"`
}

func mihomoNativeConfigPath(runtime *Runtime) (string, error) {
	if runtime == nil {
		return "", errors.New("runtime is required")
	}
	return mihomoNativeConfigPathFromDir(runtime.cfg.Mihomo.ConfigDir)
}

func mihomoNativeConfigPathFromDir(configDir string) (string, error) {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return "", errors.New("mihomo config dir is required")
	}
	return filepath.Join(configDir, mihomoNativeFileName), nil
}

func (r *Runtime) projectMihomoNativeSettings(ctx context.Context) error {
	if r == nil || r.settings == nil {
		return nil
	}
	view, err := r.settings.loadMihomoNative(ctx)
	if err != nil {
		return err
	}
	if mihomoNativeSettingsEmpty(view) {
		migrated, err := r.importMihomoNativeProjection(ctx)
		if err != nil {
			return err
		}
		if !mihomoNativeSettingsEmpty(migrated) {
			view = migrated
		}
	}
	if mihomoNativeSettingsEmpty(view) && strings.TrimSpace(r.cfg.Mihomo.ConfigDir) == "" {
		return nil
	}
	config, err := mihomoNativeConfigFileFromSettings(view)
	if err != nil {
		return err
	}
	return saveMihomoNativeConfig(r, config)
}

func (r *Runtime) importMihomoNativeProjection(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	config, exists, err := loadMihomoNativeProjection(r)
	if err != nil || !exists {
		return nil, err
	}
	view := mihomoNativeSettingsFromConfig(config)
	if mihomoNativeSettingsEmpty(view) {
		return nil, nil
	}
	if err := r.settings.saveMihomoNative(ctx, view); err != nil {
		return nil, err
	}
	return view, nil
}

func loadMihomoNativeProjection(runtime *Runtime) (mihomoNativeConfigFile, bool, error) {
	path, err := mihomoNativeConfigPath(runtime)
	if err != nil {
		return mihomoNativeConfigFile{}, false, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return mihomoNativeConfigFile{}, false, nil
	}
	if err != nil {
		return mihomoNativeConfigFile{}, false, fmt.Errorf("read mihomo native config projection: %w", err)
	}
	if len(data) == 0 {
		return mihomoNativeConfigFile{}, false, nil
	}
	var config mihomoNativeConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return mihomoNativeConfigFile{}, false, fmt.Errorf("parse mihomo native config projection: %w", err)
	}
	if config.ProxyProviders == nil {
		config.ProxyProviders = map[string]mihomoNativeProvider{}
	}
	return config, true, nil
}

func mihomoNativeSettingsEmpty(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) bool {
	return view == nil || len(view.GetFixedProxies()) == 0 && len(view.GetSubscriptions()) == 0
}

func saveMihomoNativeConfig(runtime *Runtime, config mihomoNativeConfigFile) error {
	path, err := mihomoNativeConfigPath(runtime)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create mihomo config dir: %w", err)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mihomo native config: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".native-*.json")
	if err != nil {
		return fmt.Errorf("create mihomo native config file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write mihomo native config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close mihomo native config: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return fmt.Errorf("chmod mihomo native config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace mihomo native config: %w", err)
	}
	return nil
}

func mihomoNativeSettings(ctx context.Context, runtime *Runtime) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if runtime == nil || runtime.settings == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	return runtime.settings.loadMihomoNative(ctx)
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

func updateMihomoNativeSettings(ctx context.Context, runtime *Runtime, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if runtime == nil {
		return nil, internalError("runtime is required", nil)
	}
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	currentView, err := runtime.settings.loadMihomoNative(ctx)
	if err != nil {
		return nil, internalError("load mihomo native settings", err)
	}
	current, err := mihomoNativeConfigFileFromSettings(currentView)
	if err != nil {
		return nil, internalError("load mihomo native settings", err)
	}
	next := mihomoNativeConfigFile{
		FixedProxies:   make([]mihomoNativeFixedProxy, 0, len(view.FixedProxies)),
		Proxies:        make([]map[string]any, 0, len(view.FixedProxies)),
		ProxyProviders: map[string]mihomoNativeProvider{},
		ProxyGroups:    preserveNativeGroups(current.ProxyGroups),
		Rules:          append([]string(nil), current.Rules...),
	}
	currentFixedByID, currentFixedByName := currentFixedProxyIndexes(current)
	resourceReplacements := map[string]mihomoNativeResourceReplacement{}
	seenFixedIDs := map[string]struct{}{}
	seenFixedNames := map[string]struct{}{}
	for _, item := range view.GetFixedProxies() {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(item), currentFixedByName)
		if existing := currentFixedByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			resourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		}
		resourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		if _, exists := seenFixedIDs[normalized.ID]; exists {
			return nil, invalidArgument(fmt.Sprintf("fixed proxy %q duplicates id %q", normalized.Name, normalized.ID), nil)
		}
		if _, exists := seenFixedNames[normalized.Name]; exists {
			return nil, invalidArgument(fmt.Sprintf("fixed proxy %q duplicates name", normalized.Name), nil)
		}
		seenFixedIDs[normalized.ID] = struct{}{}
		seenFixedNames[normalized.Name] = struct{}{}
		proxy, err := mihomoNativeProxyFromURI(normalized.Name, normalized.URI)
		if err != nil {
			return nil, invalidArgument(err.Error(), nil)
		}
		normalized.Type = jsonStringValue(proxy["type"])
		next.FixedProxies = append(next.FixedProxies, normalized)
		next.Proxies = append(next.Proxies, proxy)
	}
	currentSubscriptionsByID, currentSubscriptionsByName := currentSubscriptionIndexes(current)
	seenSubscriptionIDs := map[string]struct{}{}
	seenSubscriptionNames := map[string]struct{}{}
	for _, item := range view.GetSubscriptions() {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(item), currentSubscriptionsByName)
		if existing := currentSubscriptionsByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			resourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		}
		resourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		if _, exists := seenSubscriptionIDs[normalized.ID]; exists {
			return nil, invalidArgument(fmt.Sprintf("subscription %q duplicates id %q", normalized.Name, normalized.ID), nil)
		}
		if _, exists := seenSubscriptionNames[normalized.Name]; exists {
			return nil, invalidArgument(fmt.Sprintf("subscription %q duplicates name", normalized.Name), nil)
		}
		seenSubscriptionIDs[normalized.ID] = struct{}{}
		seenSubscriptionNames[normalized.Name] = struct{}{}
		provider, subscription, err := mihomoNativeSubscriptionProvider(normalized)
		if err != nil {
			return nil, err
		}
		next.Subscriptions = append(next.Subscriptions, subscription)
		next.ProxyProviders[subscription.Name] = provider
	}
	if err := runtime.settings.saveMihomoNative(ctx, mihomoNativeSettingsFromConfig(next)); err != nil {
		return nil, internalError("save mihomo native settings", err)
	}
	if err := saveMihomoNativeConfig(runtime, next); err != nil {
		return nil, internalError("save mihomo native config", err)
	}
	_, err = runtime.settings.replaceMihomoResourceRefs(ctx, resourceReplacements)
	if err != nil {
		return nil, internalError("update mihomo native resource references", err)
	}
	runtime.exitCheckCache.clear()
	runtime.requestReconcile()
	return mihomoNativeSettings(ctx, runtime)
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

func mihomoNativeProxyFromURI(name string, rawURI string) (map[string]any, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("fixed proxy name is required")
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return nil, fmt.Errorf("fixed proxy %q uri is invalid", name)
	}
	if strings.ToLower(parsed.Scheme) != "vless" {
		return nil, fmt.Errorf("fixed proxy %q only supports vless uri now", name)
	}
	host := parsed.Hostname()
	portValue := parsed.Port()
	if host == "" || portValue == "" || parsed.User == nil {
		return nil, fmt.Errorf("fixed proxy %q vless uri requires uuid, host and port", name)
	}
	port, err := strconv.Atoi(portValue)
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("fixed proxy %q has invalid port %q", name, portValue)
	}
	query := parsed.Query()
	security := strings.ToLower(strings.TrimSpace(query.Get("security")))
	network := strings.ToLower(firstNonEmpty(query.Get("type"), query.Get("network"), "tcp"))
	config := map[string]any{
		"name":       name,
		"type":       "vless",
		"server":     host,
		"port":       port,
		"uuid":       parsed.User.Username(),
		"udp":        true,
		"network":    network,
		"encryption": firstNonEmpty(query.Get("encryption"), "none"),
	}
	if flow := strings.TrimSpace(query.Get("flow")); flow != "" {
		config["flow"] = flow
	}
	if fingerprint := firstNonEmpty(query.Get("fp"), query.Get("client-fingerprint")); fingerprint != "" {
		config["client-fingerprint"] = fingerprint
	}
	if security == "tls" || security == "reality" {
		config["tls"] = true
	}
	if serverName := firstNonEmpty(query.Get("sni"), query.Get("servername")); serverName != "" {
		config["servername"] = serverName
	}
	if security == "reality" {
		reality := map[string]any{}
		if value := firstNonEmpty(query.Get("pbk"), query.Get("public-key")); value != "" {
			reality["public-key"] = value
		}
		if value := firstNonEmpty(query.Get("sid"), query.Get("short-id")); value != "" {
			reality["short-id"] = value
		}
		if len(reality) > 0 {
			config["reality-opts"] = reality
		}
	}
	switch network {
	case "ws", "websocket":
		config["network"] = "ws"
		opts := map[string]any{}
		if value := strings.TrimSpace(query.Get("path")); value != "" {
			opts["path"] = value
		}
		if value := strings.TrimSpace(query.Get("host")); value != "" {
			opts["headers"] = map[string]string{"Host": value}
		}
		if len(opts) > 0 {
			config["ws-opts"] = opts
		}
	case "grpc":
		opts := map[string]any{}
		if value := firstNonEmpty(query.Get("serviceName"), query.Get("service-name")); value != "" {
			opts["grpc-service-name"] = value
		}
		if len(opts) > 0 {
			config["grpc-opts"] = opts
		}
	case "tcp":
	default:
		return nil, fmt.Errorf("fixed proxy %q has unsupported vless network %q", name, network)
	}
	return config, nil
}

func mihomoNativeProxyURI(proxy map[string]any) string {
	if strings.ToLower(jsonStringValue(proxy["type"])) != "vless" {
		return ""
	}
	uuid := jsonStringValue(proxy["uuid"])
	server := jsonStringValue(proxy["server"])
	port := jsonIntValue(proxy["port"])
	if uuid == "" || server == "" || port <= 0 {
		return ""
	}
	query := url.Values{}
	if network := jsonStringValue(proxy["network"]); network != "" && network != "tcp" {
		query.Set("type", network)
	}
	if encryption := jsonStringValue(proxy["encryption"]); encryption != "" {
		query.Set("encryption", encryption)
	}
	if boolValue(proxy["tls"]) {
		query.Set("security", "tls")
	}
	if serverName := jsonStringValue(proxy["servername"]); serverName != "" {
		query.Set("sni", serverName)
	}
	if fingerprint := jsonStringValue(proxy["client-fingerprint"]); fingerprint != "" {
		query.Set("fp", fingerprint)
	}
	if flow := jsonStringValue(proxy["flow"]); flow != "" {
		query.Set("flow", flow)
	}
	addRealityQuery(query, proxy["reality-opts"])
	addWSQuery(query, proxy["ws-opts"])
	addGRPCQuery(query, proxy["grpc-opts"])
	uri := url.URL{Scheme: "vless", User: url.User(uuid), Host: net.JoinHostPort(server, strconv.Itoa(port)), RawQuery: query.Encode(), Fragment: jsonStringValue(proxy["name"])}
	return uri.String()
}

func addRealityQuery(query url.Values, value any) {
	opts, ok := value.(map[string]any)
	if !ok {
		return
	}
	if publicKey := jsonStringValue(opts["public-key"]); publicKey != "" {
		query.Set("security", "reality")
		query.Set("pbk", publicKey)
	}
	if shortID := jsonStringValue(opts["short-id"]); shortID != "" {
		query.Set("sid", shortID)
	}
}

func addWSQuery(query url.Values, value any) {
	opts, ok := value.(map[string]any)
	if !ok {
		return
	}
	if path := jsonStringValue(opts["path"]); path != "" {
		query.Set("path", path)
	}
	headers, ok := opts["headers"].(map[string]any)
	if ok {
		if host := jsonStringValue(headers["Host"]); host != "" {
			query.Set("host", host)
		}
	}
}

func addGRPCQuery(query url.Values, value any) {
	opts, ok := value.(map[string]any)
	if !ok {
		return
	}
	if serviceName := jsonStringValue(opts["grpc-service-name"]); serviceName != "" {
		query.Set("serviceName", serviceName)
	}
}

func jsonStringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func jsonIntValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	default:
		return 0
	}
}

func boolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return false
	}
}
