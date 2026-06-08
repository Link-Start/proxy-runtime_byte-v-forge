package runtimehttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	xproxy "golang.org/x/net/proxy"
)

var HTTPProxySchemes = []string{"http", "https"}
var CommonProxySchemes = []string{"http", "https", "socks5", "socks5h"}

func New(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: normalizeTimeout(timeout)}
}

func NewWithProxy(timeout time.Duration, proxyRawURL string, schemes ...string) (*http.Client, error) {
	transport, err := transportWithProxy(proxyRawURL, schemes...)
	if err != nil {
		return nil, err
	}
	return &http.Client{Timeout: normalizeTimeout(timeout), Transport: transport}, nil
}

func transportWithProxy(proxyRawURL string, schemes ...string) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	proxyRawURL = strings.TrimSpace(proxyRawURL)
	if proxyRawURL == "" {
		return transport, nil
	}
	parsed, err := url.Parse(proxyRawURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy_url: %w", err)
	}
	allowed := normalizedProxySchemes(schemes)
	scheme := strings.ToLower(parsed.Scheme)
	switch scheme {
	case "http", "https":
		if !allowed[scheme] {
			return nil, unsupportedProxyScheme(scheme, allowed)
		}
		transport.Proxy = http.ProxyURL(parsed)
	case "socks5", "socks5h":
		if !allowed[scheme] {
			return nil, unsupportedProxyScheme(scheme, allowed)
		}
		var auth *xproxy.Auth
		if parsed.User != nil {
			password, _ := parsed.User.Password()
			auth = &xproxy.Auth{User: parsed.User.Username(), Password: password}
		}
		dialer, err := xproxy.SOCKS5("tcp", parsed.Host, auth, xproxy.Direct)
		if err != nil {
			return nil, err
		}
		transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}
	default:
		return nil, unsupportedProxyScheme(scheme, allowed)
	}
	return transport, nil
}

func normalizedProxySchemes(values []string) map[string]bool {
	if len(values) == 0 {
		values = CommonProxySchemes
	}
	out := make(map[string]bool, len(values))
	for _, value := range values {
		if value = strings.ToLower(strings.TrimSpace(value)); value != "" {
			out[value] = true
		}
	}
	return out
}

func unsupportedProxyScheme(scheme string, allowed map[string]bool) error {
	items := make([]string, 0, len(allowed))
	for item := range allowed {
		items = append(items, item)
	}
	sort.Strings(items)
	return fmt.Errorf("proxy_url: unsupported proxy scheme %q; supported schemes: %s", scheme, strings.Join(items, ", "))
}

func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 30 * time.Second
	}
	return timeout
}
