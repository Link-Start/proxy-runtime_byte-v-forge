package app

import (
	"context"
	"fmt"
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
	settings.DynamicIpProviders = make([]*proxyruntimev1.ProxyDynamicIPProviderSettings, 0, len(providers))
	seen := map[string]struct{}{}
	for index, provider := range providers {
		item := dynamicIPProviderFromProto(provider)
		if err := validateDynamicIPProvider(item, index, s.accountProviders); err != nil {
			return nil, err
		}
		id := dynamicIPProviderID(item)
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("dynamic_ip_providers[%d] duplicates dynamic provider %q", index, id)
		}
		seen[id] = struct{}{}
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
		return normalizeRuntimeSettingsWithProviders(nil, s.ipFraudProviders, s.ipGeoProviders), nil
	}
	settings, err := s.store.LoadRuntimeSettings(ctx)
	if err != nil {
		return nil, err
	}
	settings = normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders)
	if ensurePlaygroundInUserRules(settings) {
		if err := s.store.SaveRuntimeSettings(ctx, settings); err != nil {
			return nil, err
		}
	}
	return settings, nil
}

func (s *runtimeSettingsStore) saveLocked(ctx context.Context, settings *runtimeSettingsFile) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveRuntimeSettings(ctx, normalizeRuntimeSettingsWithProviders(settings, s.ipFraudProviders, s.ipGeoProviders))
}

func (s *runtimeSettingsStore) replace(ctx context.Context, settings *runtimeSettingsFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(ctx, settings)
}

func (s *runtimeSettingsStore) loadMihomoNative(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadMihomoNativeLocked(ctx)
}

func (s *runtimeSettingsStore) loadMihomoNativeLocked(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if s.store == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	settings, err := s.store.LoadMihomoNativeSettings(ctx)
	if err != nil {
		return nil, err
	}
	return normalizeMihomoNativeSettings(settings), nil
}

func (s *runtimeSettingsStore) saveMihomoNative(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveMihomoNativeLocked(ctx, settings)
}

func (s *runtimeSettingsStore) saveMihomoNativeLocked(ctx context.Context, settings *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error {
	if s.store == nil {
		return nil
	}
	return s.store.SaveMihomoNativeSettings(ctx, normalizeMihomoNativeSettings(settings))
}
