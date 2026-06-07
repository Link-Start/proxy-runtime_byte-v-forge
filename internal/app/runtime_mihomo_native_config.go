package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	mihomoNativeFileName        = "native.json"
	mihomoFixedProxyGroupName   = "固定代理"
	defaultProviderHealthURL    = "https://www.gstatic.com/generate_204"
	defaultProviderHealthPeriod = 300
	defaultProviderHealthWait   = 5000
	defaultProviderUserAgent    = "mihomo/1.18.3"
	maxSubscriptionContentBytes = 2 * 1024 * 1024
)

type mihomoNativeSettingsView struct {
	FixedProxies  []mihomoNativeFixedProxy   `json:"fixed_proxies"`
	Subscriptions []mihomoNativeSubscription `json:"subscriptions"`
}

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
	Extra          map[string]json.RawMessage      `json:"-"`
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
	configDir := strings.TrimSpace(runtime.cfg.Mihomo.ConfigDir)
	if configDir == "" {
		return "", errors.New("mihomo config dir is required")
	}
	return filepath.Join(configDir, mihomoNativeFileName), nil
}

func loadMihomoNativeConfig(runtime *Runtime) (mihomoNativeConfigFile, error) {
	path, err := mihomoNativeConfigPath(runtime)
	if err != nil {
		return mihomoNativeConfigFile{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return mihomoNativeConfigFile{ProxyProviders: map[string]mihomoNativeProvider{}}, nil
	}
	if err != nil {
		return mihomoNativeConfigFile{}, fmt.Errorf("read mihomo native config: %w", err)
	}
	if len(data) == 0 {
		return mihomoNativeConfigFile{ProxyProviders: map[string]mihomoNativeProvider{}}, nil
	}
	var config mihomoNativeConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return mihomoNativeConfigFile{}, fmt.Errorf("parse mihomo native config: %w", err)
	}
	if config.ProxyProviders == nil {
		config.ProxyProviders = map[string]mihomoNativeProvider{}
	}
	return config, nil
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

func mihomoNativeSettings(runtime *Runtime) (*mihomoNativeSettingsView, error) {
	config, err := loadMihomoNativeConfig(runtime)
	if err != nil {
		return nil, err
	}
	view := &mihomoNativeSettingsView{}
	if len(config.FixedProxies) > 0 {
		for _, proxy := range config.FixedProxies {
			item := normalizeMihomoNativeFixedProxy(proxy, nil)
			if item.Name == "" || item.URI == "" {
				continue
			}
			view.FixedProxies = append(view.FixedProxies, item)
		}
	} else {
		for _, proxy := range config.Proxies {
			name := jsonStringValue(proxy["name"])
			proxyType := jsonStringValue(proxy["type"])
			if name == "" {
				continue
			}
			uri := mihomoNativeProxyURI(proxy)
			view.FixedProxies = append(view.FixedProxies, mihomoNativeFixedProxy{ID: nativeStableID("fixed", uri), Name: name, Type: proxyType, URI: uri})
		}
	}
	if len(config.Subscriptions) > 0 {
		for _, subscription := range config.Subscriptions {
			item := normalizeMihomoNativeSubscription(subscription, nil)
			if item.Name == "" || item.URL == "" {
				continue
			}
			view.Subscriptions = append(view.Subscriptions, item)
		}
	} else {
		for name, provider := range config.ProxyProviders {
			if strings.TrimSpace(provider.URL) == "" {
				continue
			}
			view.Subscriptions = append(view.Subscriptions, mihomoNativeSubscription{ID: nativeStableID("sub", provider.URL), Name: name, URL: provider.URL})
		}
	}
	return view, nil
}

func updateMihomoNativeSettings(ctx context.Context, runtime *Runtime, view mihomoNativeSettingsView) (*mihomoNativeSettingsView, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	current, err := loadMihomoNativeConfig(runtime)
	if err != nil {
		return nil, err
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
	for _, item := range view.FixedProxies {
		normalized := normalizeMihomoNativeFixedProxy(item, currentFixedByName)
		if existing := currentFixedByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			resourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		}
		resourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		if _, exists := seenFixedIDs[normalized.ID]; exists {
			return nil, fmt.Errorf("fixed proxy %q duplicates id %q", normalized.Name, normalized.ID)
		}
		if _, exists := seenFixedNames[normalized.Name]; exists {
			return nil, fmt.Errorf("fixed proxy %q duplicates name", normalized.Name)
		}
		seenFixedIDs[normalized.ID] = struct{}{}
		seenFixedNames[normalized.Name] = struct{}{}
		proxy, err := mihomoNativeProxyFromURI(normalized.Name, normalized.URI)
		if err != nil {
			return nil, err
		}
		normalized.Type = jsonStringValue(proxy["type"])
		next.FixedProxies = append(next.FixedProxies, normalized)
		next.Proxies = append(next.Proxies, proxy)
	}
	providerPaths := map[string]string{}
	subscriptionURLs := map[string]string{}
	for name, provider := range current.ProxyProviders {
		providerPaths[name] = provider.Path
		if strings.TrimSpace(provider.URL) != "" {
			subscriptionURLs[name] = provider.URL
		}
	}
	for _, item := range current.Subscriptions {
		if strings.TrimSpace(item.Name) != "" {
			subscriptionURLs[item.Name] = item.URL
		}
	}
	currentSubscriptionsByID, currentSubscriptionsByName := currentSubscriptionIndexes(current)
	usedProviderPaths := map[string]struct{}{}
	seenSubscriptionIDs := map[string]struct{}{}
	seenSubscriptionNames := map[string]struct{}{}
	for _, item := range view.Subscriptions {
		normalized := normalizeMihomoNativeSubscription(item, currentSubscriptionsByName)
		if existing := currentSubscriptionsByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			resourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		}
		resourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		if _, exists := seenSubscriptionIDs[normalized.ID]; exists {
			return nil, fmt.Errorf("subscription %q duplicates id %q", normalized.Name, normalized.ID)
		}
		if _, exists := seenSubscriptionNames[normalized.Name]; exists {
			return nil, fmt.Errorf("subscription %q duplicates name", normalized.Name)
		}
		seenSubscriptionIDs[normalized.ID] = struct{}{}
		seenSubscriptionNames[normalized.Name] = struct{}{}
		provider, subscription, err := mihomoNativeSubscriptionProvider(ctx, runtime, normalized, providerPaths[normalized.Name], subscriptionURLs[normalized.Name])
		if err != nil {
			return nil, err
		}
		next.Subscriptions = append(next.Subscriptions, subscription)
		next.ProxyProviders[subscription.Name] = provider
		if strings.TrimSpace(provider.Path) != "" {
			usedProviderPaths[provider.Path] = struct{}{}
		}
	}
	next.ProxyGroups = withFixedProxyGroup(next.ProxyGroups, next.Proxies)
	if err := saveMihomoNativeConfig(runtime, next); err != nil {
		return nil, err
	}
	removeUnusedNativeProviderFiles(runtime, providerPaths, usedProviderPaths)
	if err := runtime.settings.replaceMihomoResourceRefs(ctx, resourceReplacements); err != nil {
		return nil, err
	}
	runtime.exitCheckCache.clear()
	if err := runtime.runReconcile(ctx); err != nil {
		return nil, err
	}
	return mihomoNativeSettings(runtime)
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

func withFixedProxyGroup(groups []mihomoNativeGroup, proxies []map[string]any) []mihomoNativeGroup {
	names := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		if name := jsonStringValue(proxy["name"]); name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return groups
	}
	group := mihomoNativeGroup{Name: mihomoFixedProxyGroupName, Type: "select", Proxies: names, Lazy: true}
	return append([]mihomoNativeGroup{group}, groups...)
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

func stableProviderPath(path string, id string) bool {
	path = strings.TrimSpace(path)
	id = runtimeSafeID(id)
	if path == "" || id == "" {
		return false
	}
	return filepath.Base(path) == id+".yaml"
}

func mihomoNativeSubscriptionProvider(ctx context.Context, runtime *Runtime, item mihomoNativeSubscription, existingPath string, existingURL string) (mihomoNativeProvider, mihomoNativeSubscription, error) {
	id := runtimeSafeID(item.ID)
	name := strings.TrimSpace(item.Name)
	rawURL := strings.TrimSpace(item.URL)
	if id == "" {
		id = nativeStableID("sub", rawURL)
	}
	if name == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, errors.New("subscription name is required")
	}
	if rawURL == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, fmt.Errorf("subscription %q url is required", name)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, fmt.Errorf("subscription %q url is invalid", name)
	}
	path := ""
	if stableProviderPath(existingPath, id) {
		path = strings.TrimSpace(existingPath)
	}
	if path == "" {
		path = filepath.Join("providers", id+".yaml")
	}
	if runtime != nil && strings.TrimSpace(runtime.cfg.Mihomo.ConfigDir) != "" && !filepath.IsAbs(path) {
		path = filepath.Join(runtime.cfg.Mihomo.ConfigDir, path)
	}
	if rawURL != strings.TrimSpace(existingURL) || !regularFileExists(path) {
		content, err := fetchSubscriptionContent(ctx, rawURL)
		if err != nil {
			return mihomoNativeProvider{}, mihomoNativeSubscription{}, fmt.Errorf("subscription %q fetch failed: %w", name, err)
		}
		if err := writeSubscriptionProviderFile(path, content); err != nil {
			return mihomoNativeProvider{}, mihomoNativeSubscription{}, fmt.Errorf("subscription %q write provider file failed: %w", name, err)
		}
	}
	return mihomoNativeProvider{
		Type: "file",
		Path: path,
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

func fetchSubscriptionContent(ctx context.Context, rawURL string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			DisableKeepAlives:   true,
			ForceAttemptHTTP2:   false,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", defaultProviderUserAgent)
		resp, err := client.Do(req)
		if err == nil {
			content, readErr := readSubscriptionResponse(resp)
			if readErr == nil {
				return content, nil
			}
			err = readErr
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 300 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func readSubscriptionResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstream returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSubscriptionContentBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSubscriptionContentBytes {
		return nil, fmt.Errorf("content exceeds %d bytes", maxSubscriptionContentBytes)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, errors.New("content is empty")
	}
	return data, nil
}

func writeSubscriptionProviderFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".provider-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func removeUnusedNativeProviderFiles(runtime *Runtime, existing map[string]string, used map[string]struct{}) {
	providerDir := ""
	if runtime != nil && strings.TrimSpace(runtime.cfg.Mihomo.ConfigDir) != "" {
		providerDir = filepath.Join(runtime.cfg.Mihomo.ConfigDir, "providers")
	}
	for _, path := range existing {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := used[path]; ok {
			continue
		}
		if providerDir == "" || !isPathWithin(path, providerDir) {
			continue
		}
		_ = os.Remove(path)
	}
}

func isPathWithin(path string, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}

func mihomoNativeProxyFromURI(name string, rawURI string) (map[string]any, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("fixed proxy name is required")
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return nil, fmt.Errorf("parse fixed proxy uri: %w", err)
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

func nativeConfigID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var out strings.Builder
	lastDash := false
	for _, r := range value {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if ok {
			out.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			out.WriteByte('-')
			lastDash = true
		}
	}
	id := strings.Trim(out.String(), "-")
	if id == "" {
		return "provider"
	}
	return id
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
