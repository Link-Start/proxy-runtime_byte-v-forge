package app

import (
	"context"
	"log/slog"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
	"github.com/byte-v-forge/proxy-runtime/internal/app/settings/adapter/persistence"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings/application"
	"github.com/byte-v-forge/proxy-runtime/internal/config"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

type runtimeSettingsApplicationDependencies struct {
	Logger                     *slog.Logger
	Settings                   *persistence.Store
	ProxyUsers                 []config.ProxyUserRoute
	IPFraudProviderViews       func() []*proxyruntimev1.ProxyIPFraudProviderDescriptor
	IPGeoProviderViews         func() []*proxyruntimev1.ProxyIPGeoProviderDescriptor
	LoadMihomoNativeSettings   func(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	UpdateMihomoNativeSettings func(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	ScheduleApply              func([]string)
}

func newRuntimeSettingsApplication(deps runtimeSettingsApplicationDependencies) settingsapp.Application {
	return settingsapp.NewApplication(settingsapp.Dependencies{
		Repository:                  deps.Settings,
		ScheduleApply:               deps.ScheduleApply,
		Logger:                      deps.Logger,
		ProxyUsers:                  append([]config.ProxyUserRoute(nil), deps.ProxyUsers...),
		ProfileValidationError:      func(message string) error { return appcore.FailedPrecondition(message, nil) },
		IPFraudProviderViews:        deps.IPFraudProviderViews,
		IPGeoProviderViews:          deps.IPGeoProviderViews,
		LoadMihomoNativeSettings:    deps.LoadMihomoNativeSettings,
		UpdateMihomoNativeSettings:  deps.UpdateMihomoNativeSettings,
		DefaultMihomoNativeSettings: func() *proxyruntimev1.ProxyRuntimeMihomoNativeConfig { return mihomonative.NormalizeSettings(nil) },
	})
}
