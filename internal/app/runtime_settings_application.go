package app

import (
	"context"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type runtimeSettingsApplication struct {
	runtime *Runtime
}

func newRuntimeSettingsApplication(runtime *Runtime) runtimeSettingsApplication {
	return runtimeSettingsApplication{runtime: runtime}
}

func (s *RuntimeService) ListProxyIPFraudProviders(ctx context.Context, _ *proxyruntimev1.ListProxyIPFraudProvidersRequest) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return s.settings.ListProxyIPFraudProviders(ctx)
}

func (s *RuntimeService) ListProxyIPGeoProviders(ctx context.Context, _ *proxyruntimev1.ListProxyIPGeoProvidersRequest) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return s.settings.ListProxyIPGeoProviders(ctx)
}

func (s *RuntimeService) GetProxyRuntimeSettings(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeSettingsRequest) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	return s.settings.GetProxyRuntimeSettings(ctx)
}

func (s *RuntimeService) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyRuntimeSettings(ctx, req)
}

func (s *RuntimeService) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyDynamicIPProviders(ctx, req)
}

func (s *RuntimeService) UpdateProxyInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyInUserRules(ctx, req)
}

func (a runtimeSettingsApplication) ListProxyIPFraudProviders(context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPFraudProvidersResponse{Providers: a.runtime.ipFraudProviders.ProviderDescriptors()}, nil
}

func (a runtimeSettingsApplication) ListProxyIPGeoProviders(context.Context) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPGeoProvidersResponse{Providers: a.runtime.ipGeoProviders.ProviderDescriptors()}, nil
}

func (a runtimeSettingsApplication) GetProxyRuntimeSettings(ctx context.Context) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	settings, err := a.runtime.settings.view(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.runtime.cfg.ProxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.runtime.settings.update(ctx, req)
	if err != nil {
		return nil, err
	}
	a.runtime.resetIPFraudChecker()
	a.runtime.geoCache.clear()
	a.runtime.requestReconcile()
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	settings, err := a.runtime.settings.updateDynamicIPProviders(ctx, req.GetDynamicIpProviders())
	if err != nil {
		return nil, err
	}
	a.runtime.requestReconcile()
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyInUserRules(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.runtime.cfg.ProxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.runtime.settings.updateInUserRules(ctx, req.GetEgressProfiles(), req.GetIngressRules())
	if err != nil {
		return nil, err
	}
	if err := a.runtime.runReconcile(ctx); err != nil {
		return nil, err
	}
	return &proxyruntimev1.UpdateProxyRuntimeSettingsResponse{Settings: settings}, nil
}

func rejectMissingProxyUserProfiles(users []config.ProxyUserRoute, profiles []*proxyruntimev1.EgressProfileSettings) error {
	referenced := map[string]struct{}{}
	for _, user := range users {
		if strings.TrimSpace(user.Route) != config.ListenerRouteProfile {
			continue
		}
		if id := runtimeSafeID(user.ProfileID); id != "" {
			referenced[id] = struct{}{}
		}
	}
	if len(referenced) == 0 {
		return nil
	}
	enabled := enabledEgressProfileIDsFromProfiles(profiles)
	for id := range referenced {
		if _, exists := enabled[id]; !exists {
			return failedPrecondition(fmt.Sprintf("proxy user profile %q is not enabled", id), nil)
		}
	}
	return nil
}
