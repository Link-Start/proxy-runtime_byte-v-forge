package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	mihomoNativeFileName        = "native.json"
	mihomoFixedProxyGroupName   = "固定代理"
	defaultProviderHealthURL    = "https://www.gstatic.com/generate_204"
	defaultProviderInterval     = 3600
	defaultProviderHealthPeriod = 300
	defaultProviderHealthWait   = 5000
)

type mihomoNativeSettingsView struct {
	FixedProxies  []mihomoNativeFixedProxy   `json:"fixed_proxies"`
	Subscriptions []mihomoNativeSubscription `json:"subscriptions"`
}

type mihomoNativeFixedProxy struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	URI  string `json:"uri"`
}

type mihomoNativeSubscription struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type mihomoNativeConfigFile struct {
	Proxies        []map[string]any                `json:"proxies,omitempty"`
	ProxyProviders map[string]mihomoNativeProvider `json:"proxy-providers,omitempty"`
	ProxyGroups    []mihomoNativeGroup             `json:"proxy-groups,omitempty"`
	Rules          []string                        `json:"rules,omitempty"`
	Extra          map[string]json.RawMessage      `json:"-"`
}

type mihomoNativeProvider struct {
	Type        string                   `json:"type"`
	URL         string                   `json:"url"`
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
	for _, proxy := range config.Proxies {
		name := jsonStringValue(proxy["name"])
		proxyType := jsonStringValue(proxy["type"])
		if name == "" {
			continue
		}
		view.FixedProxies = append(view.FixedProxies, mihomoNativeFixedProxy{Name: name, Type: proxyType, URI: mihomoNativeProxyURI(proxy)})
	}
	for name, provider := range config.ProxyProviders {
		if strings.TrimSpace(provider.URL) == "" {
			continue
		}
		view.Subscriptions = append(view.Subscriptions, mihomoNativeSubscription{Name: name, URL: provider.URL})
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
		Proxies:        make([]map[string]any, 0, len(view.FixedProxies)),
		ProxyProviders: map[string]mihomoNativeProvider{},
		ProxyGroups:    preserveNativeGroups(current.ProxyGroups),
		Rules:          append([]string(nil), current.Rules...),
	}
	for _, item := range view.FixedProxies {
		proxy, err := mihomoNativeProxyFromURI(item.Name, item.URI)
		if err != nil {
			return nil, err
		}
		next.Proxies = append(next.Proxies, proxy)
	}
	providerPaths := map[string]string{}
	for name, provider := range current.ProxyProviders {
		providerPaths[name] = provider.Path
	}
	for _, item := range view.Subscriptions {
		provider, err := mihomoNativeSubscriptionProvider(runtime, item, providerPaths[item.Name])
		if err != nil {
			return nil, err
		}
		next.ProxyProviders[item.Name] = provider
	}
	next.ProxyGroups = withFixedProxyGroup(next.ProxyGroups, next.Proxies)
	if err := saveMihomoNativeConfig(runtime, next); err != nil {
		return nil, err
	}
	runtime.requestReconcile()
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

func mihomoNativeSubscriptionProvider(runtime *Runtime, item mihomoNativeSubscription, existingPath string) (mihomoNativeProvider, error) {
	name := strings.TrimSpace(item.Name)
	rawURL := strings.TrimSpace(item.URL)
	if name == "" {
		return mihomoNativeProvider{}, errors.New("subscription name is required")
	}
	if rawURL == "" {
		return mihomoNativeProvider{}, fmt.Errorf("subscription %q url is required", name)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return mihomoNativeProvider{}, fmt.Errorf("subscription %q url is invalid", name)
	}
	path := strings.TrimSpace(existingPath)
	if path == "" {
		path = filepath.Join("providers", nativeConfigID(name)+".yaml")
	}
	if runtime != nil && strings.TrimSpace(runtime.cfg.Mihomo.ConfigDir) != "" && !filepath.IsAbs(path) {
		path = filepath.Join(runtime.cfg.Mihomo.ConfigDir, path)
	}
	return mihomoNativeProvider{
		Type:     "http",
		URL:      rawURL,
		Path:     path,
		Interval: defaultProviderInterval,
		HealthCheck: &mihomoNativeHealthCheck{
			Enable:         true,
			URL:            defaultProviderHealthURL,
			Interval:       defaultProviderHealthPeriod,
			Timeout:        defaultProviderHealthWait,
			Lazy:           true,
			ExpectedStatus: 204,
		},
	}, nil
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
