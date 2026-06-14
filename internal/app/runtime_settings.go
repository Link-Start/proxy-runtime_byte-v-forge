package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

type runtimeSettingsFile = proxyruntimev1.ProxyRuntimePersistentSettings

const defaultProxyExitIPTimeout = 5 * time.Second

type runtimeSettingsStore struct {
	store            runtimeSettingsPersistence
	secretWriter     secretref.Writer
	accountProviders *providerregistry.Registry
	ipFraudProviders *ipfraud.Registry
	ipGeoProviders   *ipgeo.Registry
	logger           *slog.Logger
	mu               sync.Mutex
}

func newRuntimeSettingsStore(stores *RuntimeStores, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry, logger *slog.Logger) *runtimeSettingsStore {
	if logger == nil {
		logger = slog.Default()
	}
	var store runtimeSettingsPersistence
	var secretWriter secretref.Writer
	if stores != nil {
		store = stores.runtimeSettingsPersistence
		secretWriter = stores.secretStore
	}
	return &runtimeSettingsStore{store: store, secretWriter: secretWriter, accountProviders: accountProviders, ipFraudProviders: ipFraudProviders, ipGeoProviders: ipGeoProviders, logger: logger}
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
	nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := settingsFromRequest(ctx, s.secretWriter, req, current, s.accountProviders, s.ipFraudProviders, s.ipGeoProviders, nativeResourceIDs)
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
	settings.DynamicIpProviders, err = dynamicIPProvidersFromRequest(providers, s.accountProviders)
	if err != nil {
		return nil, err
	}
	if err := s.saveLocked(ctx, settings); err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}
