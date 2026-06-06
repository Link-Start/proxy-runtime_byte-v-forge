package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/proxyurl"
)

func (c Config) validate() error {
	if c.RuntimeAddr == "" {
		return errors.New("PROXY_RUNTIME_ADDR is required")
	}
	if strings.TrimSpace(c.PostgresDSN) == "" {
		return errors.New("PROXY_RUNTIME_POSTGRES_DSN or PG_DSN is required")
	}
	if strings.TrimSpace(c.EncryptionKey) == "" {
		return errors.New("PROXY_RUNTIME_ENCRYPTION_KEY is required")
	}
	if strings.TrimSpace(c.RedisURL) == "" {
		return errors.New("PLATFORM_REDIS_URL is required")
	}
	if err := c.Mihomo.validate(); err != nil {
		return err
	}
	if !isLocalProtocol(c.LocalProtocol) {
		return fmt.Errorf("unsupported local protocol %q", c.LocalProtocol)
	}
	if strings.TrimSpace(c.ProviderHTTPProxy) != "" {
		if _, err := proxyurl.Parse(c.ProviderHTTPProxy, "http"); err != nil {
			return errors.New("PROXY_RUNTIME_PROVIDER_HTTP_PROXY is invalid")
		}
	}
	if err := validateProxyUsers(c.ProxyUsers); err != nil {
		return err
	}
	if c.Provider != ProviderTen24 && c.Provider != ProviderNone {
		return ErrUnsupportedProvider
	}
	if c.Provider == ProviderTen24 && c.Ten24.HasRuntimeConfig() {
		if err := c.Ten24.Validate(); err != nil {
			return err
		}
	}
	if c.RefreshInterval < 0 {
		return errors.New("PROXY_RUNTIME_REFRESH_SECONDS must be >= 0")
	}
	if c.RequestTimeout <= 0 {
		return errors.New("PROXY_RUNTIME_REQUEST_TIMEOUT_SECONDS must be > 0")
	}
	if len(c.ProxyExitGeoURLs) == 0 {
		return errors.New("PROXY_RUNTIME_PROXY_EXIT_GEO_URLS must not be empty")
	}
	if c.EdgeCanaryTimeout <= 0 {
		return errors.New("PROXY_RUNTIME_EDGE_CANARY_TIMEOUT_SECONDS must be > 0")
	}
	if err := c.IPFraud.validate(); err != nil {
		return err
	}
	return nil
}
