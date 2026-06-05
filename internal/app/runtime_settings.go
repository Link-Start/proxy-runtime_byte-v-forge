package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

type runtimeSettingsFile = proxyruntimev1.ProxyRuntimePersistentSettings

const defaultProxyExitIPTimeout = 5 * time.Second

type runtimeSettingsStore struct {
	store            *PostgresStore
	accountProviders *providerregistry.Registry
	ipFraudProviders *ipfraud.Registry
	logger           *slog.Logger
	mu               sync.Mutex
}

func newRuntimeSettingsStore(store *PostgresStore, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, logger *slog.Logger) *runtimeSettingsStore {
	if logger == nil {
		logger = slog.Default()
	}
	return &runtimeSettingsStore{store: store, accountProviders: accountProviders, ipFraudProviders: ipFraudProviders, logger: logger}
}

func (s *runtimeSettingsStore) view(ctx context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	settings, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) update(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	sourceIDs, err := s.enabledSourceIDs(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := settingsFromRequest(ctx, s.store, req, current, s.accountProviders, s.ipFraudProviders, sourceIDs)
	if err != nil {
		return nil, err
	}
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) updateDynamicIPProviders(ctx context.Context, providers []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	settings.DynamicIpProviders = make([]*proxyruntimev1.ProxyDynamicIPProviderSettings, 0, len(providers))
	seen := map[string]struct{}{}
	for index, provider := range providers {
		item := dynamicIPProviderFromProto(provider)
		if err := validateDynamicIPProvider(item, index, s.accountProviders); err != nil {
			return nil, err
		}
		if _, exists := seen[item.GetProviderId()]; exists {
			return nil, fmt.Errorf("dynamic_ip_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seen[item.GetProviderId()] = struct{}{}
		settings.DynamicIpProviders = append(settings.DynamicIpProviders, item)
	}
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) load(ctx context.Context) (*runtimeSettingsFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked(ctx)
}

func (s *runtimeSettingsStore) loadLocked(ctx context.Context) (*runtimeSettingsFile, error) {
	if s.store == nil {
		return normalizeRuntimeSettingsWithProviders(nil, s.ipFraudProviders), nil
	}
	settings, err := s.store.LoadRuntimeSettings(ctx)
	if err != nil {
		return nil, err
	}
	return normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders), nil
}

func (s *runtimeSettingsStore) saveLocked(ctx context.Context, settings *runtimeSettingsFile) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveRuntimeSettings(ctx, normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders))
}
