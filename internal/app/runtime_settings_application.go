package app

import (
	"context"
	"log/slog"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type runtimeSettingsApplicationDependencies struct {
	Logger                     *slog.Logger
	Settings                   runtimeSettingsRepository
	ProxyUsers                 []config.ProxyUserRoute
	IPFraudProviderViews       func() []*proxyruntimev1.ProxyIPFraudProviderDescriptor
	IPGeoProviderViews         func() []*proxyruntimev1.ProxyIPGeoProviderDescriptor
	LoadMihomoNativeSettings   func(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	UpdateMihomoNativeSettings func(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	ScheduleApply              func([]string)
}

func newRuntimeSettingsApplication(deps runtimeSettingsApplicationDependencies) settingsapp.Application {
	return settingsapp.NewApplication(settingsapp.Dependencies{
		Repository:                  runtimeSettingsRepositoryAdapter{repository: deps.Settings},
		ScheduleApply:               deps.ScheduleApply,
		Logger:                      deps.Logger,
		ProxyUsers:                  append([]config.ProxyUserRoute(nil), deps.ProxyUsers...),
		ProfileValidationError:      func(message string) error { return failedPrecondition(message, nil) },
		IPFraudProviderViews:        deps.IPFraudProviderViews,
		IPGeoProviderViews:          deps.IPGeoProviderViews,
		LoadMihomoNativeSettings:    deps.LoadMihomoNativeSettings,
		UpdateMihomoNativeSettings:  deps.UpdateMihomoNativeSettings,
		DefaultMihomoNativeSettings: func() *proxyruntimev1.ProxyRuntimeMihomoNativeConfig { return normalizeMihomoNativeSettings(nil) },
	})
}
