package app

import (
	"context"
	"log/slog"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"github.com/byte-v-forge/proxy-gateway/internal/config"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"
)

type ipFraudChecker interface {
	Check(ctx context.Context, ip string) (*proxygatewayv1.ProxyIPFraudCheck, error)
}

func newIPFraudChecker(registry *ipfraud.Registry, cfg config.IPFraudConfig, providers []ipfraud.ProviderConfig, logger *slog.Logger, clk clock.Clock) ipFraudChecker {
	return ipfraud.NewService(registry, ipfraud.Config{
		Providers:   providers,
		Timeout:     cfg.Timeout,
		CacheTTL:    cfg.CacheTTL,
		KeyCooldown: cfg.KeyCooldown,
		Clock:       clk,
	}, logger)
}
