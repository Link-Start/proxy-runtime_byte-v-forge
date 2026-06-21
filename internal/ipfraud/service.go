package ipfraud

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"github.com/byte-v-forge/proxy-gateway/internal/runtimehttp"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	providers []providerEntry
	cacheTTL  time.Duration
	logger    *slog.Logger
	clock     clock.Clock

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	check     *proxygatewayv1.ProxyIPFraudCheck
	expiresAt time.Time
}

type providerEntry struct {
	id          string
	displayName string
	checker     provider
}

func NewService(registry *Registry, cfg Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 10 * time.Minute
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.SystemClock{}
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
		providers = append(providers, providerEntry{
			id:          providerID,
			displayName: plugin.DisplayName(),
			checker:     plugin.New(client, item, cfg.KeyCooldown, cfg.Clock),
		})
	}
	return &Service{
		providers: providers,
		cacheTTL:  cfg.CacheTTL,
		logger:    logger,
		clock:     cfg.Clock,
		cache:     map[string]cacheEntry{},
	}
}

func (s *Service) Check(ctx context.Context, ip string) (*proxygatewayv1.ProxyIPFraudCheck, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil, errors.New("ip is required")
	}
	if cached := s.cached(ip, s.clock.Now()); cached != nil {
		return cached, nil
	}
	reports := make([]report, 0, len(s.providers))
	for _, item := range s.providers {
		result, err := item.checker.lookup(ctx, ip)
		if err != nil {
			s.logger.Warn("IP fraud provider unavailable", "provider", item.id, "error_type", fmt.Sprintf("%T", err))
			continue
		}
		result.providerID = item.id
		result.providerName = item.displayName
		reports = append(reports, result)
	}
	errorMessage := ""
	if len(reports) == 0 {
		errorMessage = "IP fraud check unavailable"
	}
	check := mergeReports(ip, reports, errorMessage)
	s.store(ip, check, s.clock.Now().Add(s.cacheTTL))
	return cloneCheck(check), nil
}

func (s *Service) cached(ip string, now time.Time) *proxygatewayv1.ProxyIPFraudCheck {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[ip]
	if !ok {
		return nil
	}
	if !entry.expiresAt.After(now) {
		delete(s.cache, ip)
		return nil
	}
	return cloneCheck(entry.check)
}

func (s *Service) store(ip string, check *proxygatewayv1.ProxyIPFraudCheck, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[ip] = cacheEntry{check: cloneCheck(check), expiresAt: expiresAt}
}

func cloneCheck(in *proxygatewayv1.ProxyIPFraudCheck) *proxygatewayv1.ProxyIPFraudCheck {
	if in == nil {
		return nil
	}
	return proto.Clone(in).(*proxygatewayv1.ProxyIPFraudCheck)
}
