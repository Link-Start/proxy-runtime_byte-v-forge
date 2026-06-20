package app

import (
	"context"
	"log/slog"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
)

type ipFraudChecker interface {
	Check(ctx context.Context, ip string) (*proxyruntimev1.ProxyIPFraudCheck, error)
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
