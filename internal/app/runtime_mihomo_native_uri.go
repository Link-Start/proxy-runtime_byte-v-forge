package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

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
