package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/proxyurl"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
)

func shortHash(value string) string {
	h := hashModulo(value, 0xffffffff)
	return fmt.Sprintf("%08x", h)
}

func hashModulo(value string, modulo uint32) uint32 {
	var h uint32 = 2166136261
	for _, ch := range []byte(value) {
		h ^= uint32(ch)
		h *= 16777619
	}
	if modulo > 0 {
		return h % modulo
	}
	return h
}

func cloneLabels(labels map[string]string) map[string]string {
	cloned := map[string]string{}
	for k, v := range labels {
		cloned[k] = v
	}
	return cloned
}

func cloneNodes(in []provider.Node) []provider.Node {
	out := make([]provider.Node, 0, len(in))
	for _, node := range in {
		cloned := node
		if node.URL != nil {
			u := *node.URL
			cloned.URL = &u
		}
		cloned.Labels = cloneLabels(node.Labels)
		out = append(out, cloned)
	}
	return out
}

func NewProviderHTTPClient(cfg config.Config) (*http.Client, error) {
	proxyURL := ""
	if strings.TrimSpace(cfg.ProviderHTTPProxy) != "" {
		parsed, err := proxyurl.Parse(cfg.ProviderHTTPProxy, "http")
		if err != nil {
			return nil, fmt.Errorf("parse provider HTTP proxy: %w", err)
		}
		proxyURL = parsed.String()
	}
	client, err := runtimehttp.NewWithProxy(cfg.RequestTimeout, proxyURL, runtimehttp.HTTPProxySchemes...)
	if err != nil {
		return nil, err
	}
	return client, nil
}
