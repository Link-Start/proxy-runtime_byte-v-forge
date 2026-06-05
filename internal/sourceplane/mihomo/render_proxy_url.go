package mihomo

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func renderProviderNodes(prefix string, nodes []provider.Node) ([]map[string]any, []string, error) {
	proxies := make([]map[string]any, 0, len(nodes))
	names := make([]string, 0, len(nodes))
	for index, node := range nodes {
		name := providerNodeName(prefix, node, index)
		proxy, err := renderProxyURL(name, node.URL)
		if err != nil {
			return nil, nil, fmt.Errorf("render proxy node %q: %w", name, err)
		}
		proxies = append(proxies, proxy)
		names = append(names, name)
	}
	return proxies, names, nil
}

func providerNodeName(prefix string, node provider.Node, index int) string {
	name := safeID(firstNonEmpty(node.ID, fmt.Sprintf("%s-%d", prefix, index)))
	if name == "" {
		return fmt.Sprintf("%s-%d", prefix, index)
	}
	return name
}

func renderProxyURL(name string, proxyURL *url.URL) (map[string]any, error) {
	if proxyURL == nil || strings.TrimSpace(proxyURL.Host) == "" {
		return nil, errors.New("proxy url host is required")
	}
	scheme := strings.ToLower(strings.TrimSpace(proxyURL.Scheme))
	if scheme == "" {
		scheme = "http"
	}
	host := proxyURL.Hostname()
	port, err := proxyPort(proxyURL)
	if err != nil {
		return nil, err
	}
	proxyType := scheme
	config := map[string]any{
		"name":   name,
		"type":   proxyType,
		"server": host,
		"port":   port,
		"udp":    true,
	}
	switch scheme {
	case "http":
	case "https":
		config["type"] = "http"
		config["tls"] = true
	case "socks5", "socks5h":
		config["type"] = "socks5"
	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q", scheme)
	}
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		config["username"] = proxyURL.User.Username()
		config["password"] = password
	}
	return config, nil
}

func proxyPort(proxyURL *url.URL) (int, error) {
	if proxyURL == nil {
		return 0, errors.New("proxy url is required")
	}
	if portValue := strings.TrimSpace(proxyURL.Port()); portValue != "" {
		port, err := strconv.Atoi(portValue)
		if err != nil || port <= 0 || port > 65535 {
			return 0, fmt.Errorf("invalid proxy port %q", portValue)
		}
		return port, nil
	}
	_, portValue, err := net.SplitHostPort(proxyURL.Host)
	if err == nil {
		port, parseErr := strconv.Atoi(portValue)
		if parseErr == nil && port > 0 && port <= 65535 {
			return port, nil
		}
	}
	switch strings.ToLower(strings.TrimSpace(proxyURL.Scheme)) {
	case "http":
		return 80, nil
	case "https":
		return 443, nil
	default:
		return 0, errors.New("proxy port is required")
	}
}
