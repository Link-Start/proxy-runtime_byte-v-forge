package app

import (
	"context"
	"log/slog"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"
	"github.com/byte-v-forge/proxy-gateway/internal/app/proxycheck"
	"github.com/byte-v-forge/proxy-gateway/internal/app/settings/adapter/persistence"
	settingsapp "github.com/byte-v-forge/proxy-gateway/internal/app/settings/application"
	"github.com/byte-v-forge/proxy-gateway/internal/config"
)

// runtimeSettingsFile aliases the persisted runtime-settings proto for the many
// composition-root sites that pass it around; the settings persistence adapter
// lives in internal/app/settings/adapter/persistence.
type runtimeSettingsFile = proxygatewayv1.ProxyGatewayPersistentSettings

type runtimeSettingsApplicationDependencies struct {
	Logger                     *slog.Logger
	Settings                   *persistence.Store
	ProxyUsers                 []config.ProxyUserRoute
	IPFraudProviderViews       func() []*proxygatewayv1.ProxyIPFraudProviderDescriptor
	IPGeoProviderViews         func() []*proxygatewayv1.ProxyIPGeoProviderDescriptor
	LoadMihomoNativeSettings   func(context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error)
	UpdateMihomoNativeSettings func(context.Context, *proxygatewayv1.ProxyGatewayMihomoNativeConfig) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error)
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
		DefaultMihomoNativeSettings: func() *proxygatewayv1.ProxyGatewayMihomoNativeConfig { return mihomonative.NormalizeSettings(nil) },
	})
}

const runtimeSettingsConnectionCleanupTimeout = 10 * time.Second

type runtimeSettingsApplyScheduler struct {
	resetIPFraudChecker   func()
	geoCache              *proxycheck.IPGeoCache
	exitCheckCache        *proxycheck.ExitCheckCache
	markApplyPending      func()
	requestReconcile      func()
	closeInUserConnection leaseConnectionCleanupFunc
}

func newRuntimeSettingsApplyScheduler(runtime *Runtime) runtimeSettingsApplyScheduler {
	if runtime == nil {
		return runtimeSettingsApplyScheduler{}
	}
	return runtimeSettingsApplyScheduler{
		resetIPFraudChecker:   runtime.resetIPFraudChecker,
		geoCache:              runtime.geoCache,
		exitCheckCache:        runtime.exitCheckCache,
		markApplyPending:      runtime.markSettingsApplyPending,
		requestReconcile:      runtime.requestReconcile,
		closeInUserConnection: runtime.closeMihomoInUserConnections,
	}
}

func (s runtimeSettingsApplyScheduler) Schedule(changedUsernames []string) {
	s.clearDerivedState()
	s.requestApply()
	if len(changedUsernames) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), runtimeSettingsConnectionCleanupTimeout)
		defer cancel()
		s.closeInUserConnections(ctx, changedUsernames)
	}()
}

func (s runtimeSettingsApplyScheduler) clearDerivedState() {
	if s.resetIPFraudChecker != nil {
		s.resetIPFraudChecker()
	}
	if s.geoCache != nil {
		s.geoCache.Clear()
	}
	if s.exitCheckCache != nil {
		s.exitCheckCache.Clear()
	}
}

func (s runtimeSettingsApplyScheduler) requestApply() {
	if s.markApplyPending != nil {
		s.markApplyPending()
	}
	if s.requestReconcile != nil {
		s.requestReconcile()
	}
}

func (s runtimeSettingsApplyScheduler) closeInUserConnections(ctx context.Context, changedUsernames []string) {
	if s.closeInUserConnection != nil {
		s.closeInUserConnection(ctx, changedUsernames)
	}
}
