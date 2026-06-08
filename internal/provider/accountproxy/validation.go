package accountproxy

import (
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func validateConfig(cfg Config, definition Definition) error {
	if strings.TrimSpace(cfg.Username) == "" || strings.TrimSpace(cfg.Password) == "" {
		return errors.New("username and password are required")
	}
	if err := validateProtocol(definition.DefaultProtocol); err != nil {
		return err
	}
	for _, protocol := range definition.Protocols {
		if err := validateProtocol(protocol); err != nil {
			return err
		}
	}
	for index, gateway := range cfg.Gateways {
		if strings.TrimSpace(gateway.EndpointURL) == "" {
			return fmt.Errorf("gateways[%d].endpoint_url is required", index)
		}
	}
	return nil
}

func validateProtocol(protocol string) error {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "", "http", "socks5":
		return nil
	default:
		return fmt.Errorf("unsupported proxy protocol %q", protocol)
	}
}

func defaultProtocol(protocol string, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "http":
		return "http"
	case "socks5":
		return "socks5"
	}
	if strings.EqualFold(strings.TrimSpace(fallback), "http") {
		return "http"
	}
	return "socks5"
}

func protocolEnumWithDefault(protocol string, fallback string) proxyruntimev1.ProxyProtocol {
	if defaultProtocol(protocol, fallback) == "http" {
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP
	}
	return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
}
