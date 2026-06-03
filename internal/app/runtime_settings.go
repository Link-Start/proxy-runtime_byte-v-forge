package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type runtimeSettingsFile = proxyruntimev1.ProxyRuntimePersistentSettings

const defaultProxyExitIPTimeout = 5 * time.Second

type runtimeSettingsStore struct {
	store            *PostgresStore
	accountProviders *accountproxy.Registry
	ipFraudProviders *ipfraud.Registry
	logger           *slog.Logger
	mu               sync.Mutex
}

func newRuntimeSettingsStore(store *PostgresStore, accountProviders *accountproxy.Registry, ipFraudProviders *ipfraud.Registry, logger *slog.Logger) *runtimeSettingsStore {
	if logger == nil {
		logger = slog.Default()
	}
	return &runtimeSettingsStore{store: store, accountProviders: accountProviders, ipFraudProviders: ipFraudProviders, logger: logger}
}

func (s *runtimeSettingsStore) view(ctx context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	settings, err := s.loadContext(ctx)
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
	settings, err := settingsFromRequest(ctx, s.store, req, current, s.accountProviders, s.ipFraudProviders)
	if err != nil {
		return nil, err
	}
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) load() (*runtimeSettingsFile, error) {
	return s.loadContext(context.Background())
}

func (s *runtimeSettingsStore) loadContext(ctx context.Context) (*runtimeSettingsFile, error) {
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

func (s *runtimeSettingsStore) save(settings *runtimeSettingsFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(context.Background(), settings)
}

func (s *runtimeSettingsStore) saveLocked(ctx context.Context, settings *runtimeSettingsFile) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveRuntimeSettings(ctx, normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders))
}
