package lease

import (
	"fmt"
	"net"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	LabelProxyUsername = "proxy_username"
	LabelProxyPassword = "proxy_password"
)

func NewListenerEndpoint(listener Listener, advertisedHost string, fallbackProtocol string) (*proxyruntimev1.ProxyEndpoint, error) {
	hostPort, err := listenerHostPort(listener.Addr)
	if err != nil {
		return nil, err
	}
	host, portValue, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}
	port, err := listenerPort(portValue)
	if err != nil {
		return nil, err
	}
	if advertisedHost != "" && localEndpointHost(host) {
		host = advertisedHost
	}
	labels := cloneStringMap(listener.Labels)
	if listener.Username != "" || listener.Password != "" {
		if labels == nil {
			labels = map[string]string{}
		}
		labels[LabelProxyUsername] = listener.Username
		labels[LabelProxyPassword] = listener.Password
	}
	return &proxyruntimev1.ProxyEndpoint{Id: listener.ID, Protocol: listenerProtocol(listener, fallbackProtocol), Host: host, Port: port, Labels: labels}, nil
}

func listenerHostPort(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr, nil
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("invalid listener address")
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port), nil
}

func listenerPort(portValue string) (uint32, error) {
	var port uint32
	_, err := fmt.Sscanf(portValue, "%d", &port)
	return port, err
}

func listenerProtocol(listener Listener, fallback string) proxyruntimev1.ProxyProtocol {
	if strings.TrimSpace(listener.Protocol) == "socks5" {
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	}
	if strings.TrimSpace(fallback) == "socks5" {
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	}
	return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP
}

func localEndpointHost(host string) bool {
	switch strings.TrimSpace(strings.ToLower(host)) {
	case "127.0.0.1", "localhost", "0.0.0.0", "::1":
		return true
	default:
		return false
	}
}
