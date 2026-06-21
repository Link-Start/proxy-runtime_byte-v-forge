package app

import (
	"context"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"
	mihomoapp "github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative/application"
	"github.com/byte-v-forge/proxy-gateway/internal/app/settings/adapter/persistence"
	"github.com/byte-v-forge/proxy-gateway/internal/config"
	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
)

func sourcePlaneProxyUserRoutes(settings *runtimeSettingsFile) []dataplane.ProxyUserRoute {
	settings = kernel.NormalizeRuntimeSettings(settings)
	out := make([]dataplane.ProxyUserRoute, 0, len(settings.GetIngressRules()))
	for _, rule := range settings.GetIngressRules() {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetUsername()) == "" {
			continue
		}
		out = append(out, dataplane.ProxyUserRoute{
			ID:        rule.GetRuleId(),
			Username:  rule.GetUsername(),
			Password:  rule.GetPasswordValue(),
			Route:     config.ListenerRouteProfile,
			ProfileID: rule.GetProfileId(),
		})
	}
	return out
}

type runtimeSettingsMihomoNativeAdapter struct {
	repository *persistence.Store
	configDir  string
	apply      runtimeMihomoNativeApplyScheduler
}

func newRuntimeSettingsMihomoNativeAdapter(runtime *Runtime) runtimeSettingsMihomoNativeAdapter {
	if runtime == nil {
		return runtimeSettingsMihomoNativeAdapter{}
	}
	return runtimeSettingsMihomoNativeAdapter{
		repository: runtime.settings,
		configDir:  runtime.cfg.Mihomo.ConfigDir,
		apply:      newRuntimeMihomoNativeApplyScheduler(runtime),
	}
}

func (a runtimeSettingsMihomoNativeAdapter) Load(ctx context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	if a.repository == nil {
		return mihomonative.NormalizeSettings(nil), nil
	}
	return a.repository.LoadMihomoNative(ctx)
}

func (a runtimeSettingsMihomoNativeAdapter) Update(ctx context.Context, config *proxygatewayv1.ProxyGatewayMihomoNativeConfig) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	if a.repository == nil {
		return nil, appcore.InternalError("mihomo native settings repository is required", nil)
	}
	return mihomoapp.Update(ctx, mihomoapp.UpdateDependencies{
		ConfigDir:    a.configDir,
		LoadSettings: a.repository.LoadMihomoNative,
		SaveSettings: a.repository.SaveMihomoNative,
		ReplaceRefs:  a.repository.ReplaceMihomoResourceRefs,
		AfterApply:   a.apply.Schedule,
	}, config)
}
