package app

import (
	"context"
	"log/slog"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type runtimeSettingsApplication struct {
	logger                     *slog.Logger
	settings                   runtimeSettingsRepository
	proxyUsers                 []config.ProxyUserRoute
	ipFraudProviderViews       func() []*proxyruntimev1.ProxyIPFraudProviderDescriptor
	ipGeoProviderViews         func() []*proxyruntimev1.ProxyIPGeoProviderDescriptor
	loadMihomoNativeSettings   func(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	updateMihomoNativeSettings func(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	scheduleApply              func([]string)
}

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

func newRuntimeSettingsApplication(deps runtimeSettingsApplicationDependencies) runtimeSettingsApplication {
	return runtimeSettingsApplication{
		logger:                     deps.Logger,
		settings:                   deps.Settings,
		proxyUsers:                 append([]config.ProxyUserRoute(nil), deps.ProxyUsers...),
		ipFraudProviderViews:       deps.IPFraudProviderViews,
		ipGeoProviderViews:         deps.IPGeoProviderViews,
		loadMihomoNativeSettings:   deps.LoadMihomoNativeSettings,
		updateMihomoNativeSettings: deps.UpdateMihomoNativeSettings,
		scheduleApply:              deps.ScheduleApply,
	}
}

func (a runtimeSettingsApplication) settingsUsecase() settingsapp.Application {
	return settingsapp.NewApplication(settingsapp.Dependencies{
		Repository:    runtimeSettingsRepositoryAdapter{repository: a.settings},
		ScheduleApply: a.scheduleRuntimeSettingsApply,
		Logger:        a.logger,
		ValidateProfiles: func(profiles []*proxyruntimev1.EgressProfileSettings) error {
			return rejectMissingProxyUserProfiles(a.proxyUsers, profiles)
		},
	})
}
