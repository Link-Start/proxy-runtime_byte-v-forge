package mihomo

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

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
