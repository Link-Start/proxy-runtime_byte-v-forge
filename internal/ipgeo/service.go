package ipgeo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
)

type Config struct {
	Providers []ProviderConfig
	Timeout   time.Duration
}

type Service struct {
	providers []providerEntry
	logger    *slog.Logger
}

type providerEntry struct {
	id      string
	checker provider
}

func NewService(registry *Registry, cfg Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	sort.SliceStable(cfg.Providers, func(i, j int) bool {
		return cfg.Providers[i].Weight > cfg.Providers[j].Weight
	})
	client := runtimehttp.New(cfg.Timeout)
	providers := make([]providerEntry, 0, len(cfg.Providers))
	for _, item := range cfg.Providers {
		plugin, ok := registry.PluginForKind(item.Kind)
		if !ok {
			continue
		}
		providerID := strings.TrimSpace(item.ID)
		if providerID == "" {
			providerID = plugin.ProviderID()
		}
		providers = append(providers, providerEntry{id: providerID, checker: plugin.New(client, item)})
	}
	return &Service{providers: providers, logger: logger}
}

func (s *Service) Lookup(ctx context.Context, ip string) (*proxyruntimev1.ProxyExitGeo, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil, errors.New("ip is required")
	}
	for _, item := range s.providers {
		geo, err := item.checker.Lookup(ctx, ip)
		if err != nil {
			s.logger.Warn("IP geo provider unavailable", "provider", item.id, "error_type", fmt.Sprintf("%T", err))
			continue
		}
		geo.Ip = ip
		return geo, nil
	}
	return nil, errors.New("IP geo provider unavailable")
}
