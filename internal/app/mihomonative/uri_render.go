package mihomonative

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func proxyURI(proxy map[string]any) string {
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
	addVLESSCoreQuery(query, proxy)
	addVLESSSecurityQuery(query, proxy)
	addVLESSNetworkQuery(query, proxy)
	uri := url.URL{Scheme: "vless", User: url.User(uuid), Host: net.JoinHostPort(server, strconv.Itoa(port)), RawQuery: query.Encode(), Fragment: jsonStringValue(proxy["name"])}
	return uri.String()
}
