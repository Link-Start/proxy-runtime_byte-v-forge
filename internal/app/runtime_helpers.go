package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/proxyurl"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
	"google.golang.org/protobuf/types/known/durationpb"
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

func cleanRegionCodes(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func runtimeSafeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var out strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			out.WriteRune(r)
			continue
		}
		out.WriteByte('-')
	}
	return strings.Trim(out.String(), "-")
}

func protoDuration(value *durationpb.Duration, fallback time.Duration) time.Duration {
	if value == nil || value.AsDuration() <= 0 {
		return fallback
	}
	return value.AsDuration()
}

func defaultExpectedStatus(status uint32) uint32 {
	if status == 0 {
		return 204
	}
	return status
}
