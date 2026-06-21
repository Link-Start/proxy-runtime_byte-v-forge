package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-gateway/internal/config"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
	"github.com/byte-v-forge/proxy-gateway/internal/proxyurl"
	"github.com/byte-v-forge/proxy-gateway/internal/runtimehttp"
	"google.golang.org/protobuf/types/known/durationpb"
)

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
