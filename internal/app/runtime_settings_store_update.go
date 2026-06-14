package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsMutation func(*runtimeSettingsFile) (*runtimeSettingsFile, error)

func (s *runtimeSettingsStore) view(ctx context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	settings, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	return runtimeSettingsView(settings), nil
}

func (s *runtimeSettingsStore) update(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(current *runtimeSettingsFile) (*runtimeSettingsFile, error) {
		nativeResourceIDs, err := s.enabledMihomoResourceIDs(ctx)
		if err != nil {
			return nil, err
		}
		return settingsFromRequest(ctx, s.secretWriter, req, current, s.accountProviders, s.ipFraudProviders, s.ipGeoProviders, nativeResourceIDs)
	})
}

func (s *runtimeSettingsStore) updateDynamicIPProviders(ctx context.Context, providers []*proxyruntimev1.ProxyDynamicIPProviderSettings) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	return s.mutateRuntimeSettings(ctx, func(settings *runtimeSettingsFile) (*runtimeSettingsFile, error) {
		nextProviders, err := dynamicIPProvidersFromRequest(providers, s.accountProviders)
		if err != nil {
			return nil, err
		}
		settings.DynamicIpProviders = nextProviders
		return settings, nil
	})
}

func (s *runtimeSettingsStore) mutateRuntimeSettings(ctx context.Context, mutation runtimeSettingsMutation) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings, err := s.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	next, err := mutation(settings)
	if err != nil {
		return nil, err
	}
	if err := s.saveLocked(ctx, next); err != nil {
		return nil, err
	}
	return runtimeSettingsView(next), nil
}
