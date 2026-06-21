package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

func normalizeEndpoint(endpoint sourceplane.Endpoint) (sourceplane.Endpoint, error) {
	endpoint.Addr = strings.TrimSpace(endpoint.Addr)
	if endpoint.Addr == "" {
		endpoint.Addr = "127.0.0.1:18900"
	}
	if _, _, err := splitEndpoint(endpoint.Addr); err != nil {
		return sourceplane.Endpoint{}, err
	}
	endpoint.Protocol = strings.ToLower(strings.TrimSpace(endpoint.Protocol))
	if endpoint.Protocol == "" {
		endpoint.Protocol = "socks5"
	}
	if endpoint.Protocol != "socks5" {
		return sourceplane.Endpoint{}, fmt.Errorf("unsupported mihomo source endpoint protocol %q", endpoint.Protocol)
	}
	return endpoint, nil
}
