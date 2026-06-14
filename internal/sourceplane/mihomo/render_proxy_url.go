package mihomo

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

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
