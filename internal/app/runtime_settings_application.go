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

func (s *RuntimeService) GetProxyRuntimeSettings(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeSettingsRequest) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	return s.settings.GetProxyRuntimeSettings(ctx)
}

func (s *RuntimeService) UpdateProxyRuntimeSettings(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyRuntimeSettings(ctx, req)
}

func (s *RuntimeService) UpdateProxyDynamicIPProviders(ctx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
	return s.settings.UpdateProxyDynamicIPProviders(ctx, req)
}

func (s *RuntimeService) UpdateProxyEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	return s.settings.UpdateProxyEgressProfiles(ctx, req)
}

func (s *RuntimeService) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	return s.settings.UpdateProxyIngressRules(ctx, req)
}

func (a runtimeSettingsApplication) ListProxyIPFraudProviders(context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPFraudProvidersResponse{Providers: a.runtime.ipFraudProviders.ProviderDescriptors()}, nil
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

func (a runtimeSettingsApplication) UpdateProxyEgressProfiles(ctx context.Context, req *proxyruntimev1.UpdateProxyEgressProfilesRequest) (*proxyruntimev1.UpdateProxyEgressProfilesResponse, error) {
	if err := rejectMissingProxyUserProfiles(a.runtime.cfg.ProxyUsers, req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.runtime.settings.updateEgressProfiles(ctx, req.GetEgressProfiles())
	if err != nil {
		return nil, err
	}
	a.runtime.requestReconcile()
	return &proxyruntimev1.UpdateProxyEgressProfilesResponse{Settings: settings}, nil
}

func (a runtimeSettingsApplication) UpdateProxyIngressRules(ctx context.Context, req *proxyruntimev1.UpdateProxyIngressRulesRequest) (*proxyruntimev1.UpdateProxyIngressRulesResponse, error) {
	settings, err := a.runtime.settings.updateIngressRules(ctx, req.GetIngressRules())
	if err != nil {
		return nil, err
	}
	a.runtime.requestReconcile()
	return &proxyruntimev1.UpdateProxyIngressRulesResponse{Settings: settings}, nil
}

func rejectMissingProxyUserProfiles(users []config.ProxyUserRoute, profiles []*proxyruntimev1.EgressProfileSettings) error {
	referenced := map[string]struct{}{}
	for _, user := range users {
		if strings.TrimSpace(user.Route) != config.ListenerRouteProfile {
			continue
		}
		if id := sourceSafeID(firstNonEmpty(user.ProfileID, user.SourceID)); id != "" {
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
