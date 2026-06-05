package mihomo

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
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

func splitEndpoint(addr string) (string, int, error) {
	host, portValue, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	var port int
	if _, err := fmt.Sscanf(portValue, "%d", &port); err != nil || port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid mihomo mixed port %q", portValue)
	}
	if host == "" {
		host = "0.0.0.0"
	}
	return host, port, nil
}

func waitForEndpoint(ctx context.Context, addr string, timeout time.Duration) error {
	dialAddr, err := endpointDialAddr(addr)
	if err != nil {
		return err
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		conn, err := (&net.Dialer{Timeout: 100 * time.Millisecond}).DialContext(waitCtx, "tcp", dialAddr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("mihomo mixed listener %s is not ready: %w", addr, waitCtx.Err())
		case <-ticker.C:
		}
	}
}

func endpointDialAddr(addr string) (string, error) {
	host, port, err := splitEndpoint(addr)
	if err != nil {
		return "", err
	}
	if host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, fmt.Sprintf("%d", port)), nil
}
