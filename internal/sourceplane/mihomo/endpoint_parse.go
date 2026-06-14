package mihomo

import (
	"fmt"
	"net"
)

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
