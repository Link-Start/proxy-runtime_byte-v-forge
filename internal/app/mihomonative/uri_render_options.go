package mihomonative

import "net/url"

func addVLESSCoreQuery(query url.Values, proxy map[string]any) {
	if network := jsonStringValue(proxy["network"]); network != "" && network != "tcp" {
		query.Set("type", network)
	}
	if encryption := jsonStringValue(proxy["encryption"]); encryption != "" {
		query.Set("encryption", encryption)
	}
	if fingerprint := jsonStringValue(proxy["client-fingerprint"]); fingerprint != "" {
		query.Set("fp", fingerprint)
	}
	if flow := jsonStringValue(proxy["flow"]); flow != "" {
		query.Set("flow", flow)
	}
}

func addVLESSSecurityQuery(query url.Values, proxy map[string]any) {
	if boolValue(proxy["tls"]) {
		query.Set("security", "tls")
	}
	if serverName := jsonStringValue(proxy["servername"]); serverName != "" {
		query.Set("sni", serverName)
	}
	addRealityQuery(query, proxy["reality-opts"])
}

func addVLESSNetworkQuery(query url.Values, proxy map[string]any) {
	addWSQuery(query, proxy["ws-opts"])
	addGRPCQuery(query, proxy["grpc-opts"])
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
