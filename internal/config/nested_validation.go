package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func (c IPFraudConfig) validate() error {
	if c.Timeout <= 0 {
		return errors.New("PROXY_GATEWAY_IP_FRAUD_TIMEOUT_SECONDS must be > 0")
	}
	if c.CacheTTL <= 0 {
		return errors.New("PROXY_GATEWAY_IP_FRAUD_CACHE_TTL_SECONDS must be > 0")
	}
	if c.KeyCooldown <= 0 {
		return errors.New("PROXY_GATEWAY_IP_FRAUD_KEY_COOLDOWN_SECONDS must be > 0")
	}
	return nil
}

func (c MihomoConfig) validate() error {
	if strings.TrimSpace(c.Path) == "" {
		return errors.New("PROXY_GATEWAY_MIHOMO_PATH is required")
	}
	if strings.TrimSpace(c.APIAddr) == "" {
		return errors.New("PROXY_GATEWAY_MIHOMO_API_ADDR is required")
	}
	if err := validateLoopbackHostPort("PROXY_GATEWAY_MIHOMO_API_ADDR", c.APIAddr); err != nil {
		return err
	}
	if c.HealthCheckInterval < 0 {
		return errors.New("PROXY_GATEWAY_MIHOMO_HEALTH_CHECK_INTERVAL_SECONDS must be >= 0")
	}
	if c.HealthCheckTimeout < 0 {
		return errors.New("PROXY_GATEWAY_MIHOMO_HEALTH_CHECK_TIMEOUT_SECONDS must be >= 0")
	}
	return nil
}

func validateLoopbackHostPort(name string, addr string) error {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return fmt.Errorf("%s must be loopback host:port", name)
	}
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("%s must listen on localhost or loopback IP", name)
	}
	return nil
}
