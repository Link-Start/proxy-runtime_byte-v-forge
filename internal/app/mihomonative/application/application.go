package application

import (
	"context"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

// SettingsLoader reads the persisted mihomo native settings view.
type SettingsLoader func(context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error)

// SettingsSaver persists a mihomo native settings view.
type SettingsSaver func(context.Context, *proxygatewayv1.ProxyGatewayMihomoNativeConfig) error

// ResourceRefReplacer rewrites egress resource references after an update.
type ResourceRefReplacer func(context.Context, map[string]mihomonative.ResourceReplacement) (bool, error)

// ProjectionDependencies wires the startup projection of mihomo native settings
// onto the on-disk mihomo config.
type ProjectionDependencies struct {
	ConfigDir    string
	LoadSettings SettingsLoader
	SaveSettings SettingsSaver
}

// Project renders the persisted mihomo native settings to the config directory,
// importing an existing on-disk projection when the stored view is empty.
func Project(ctx context.Context, deps ProjectionDependencies) error {
	if deps.LoadSettings == nil {
		return nil
	}
	view, err := deps.LoadSettings(ctx)
	if err != nil {
		return err
	}
	if mihomonative.SettingsEmpty(view) {
		migrated, err := importProjection(ctx, deps, deps.ConfigDir)
		if err != nil {
			return err
		}
		if !mihomonative.SettingsEmpty(migrated) {
			view = migrated
		}
	}
	if mihomonative.SettingsEmpty(view) && strings.TrimSpace(deps.ConfigDir) == "" {
		return nil
	}
	config, err := mihomonative.ConfigFromSettings(view)
	if err != nil {
		return err
	}
	return mihomonative.SaveConfig(deps.ConfigDir, config)
}

func importProjection(ctx context.Context, deps ProjectionDependencies, configDir string) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	config, exists, err := mihomonative.LoadProjection(configDir)
	if err != nil || !exists {
		return nil, err
	}
	view := mihomonative.SettingsFromConfig(config)
	if mihomonative.SettingsEmpty(view) {
		return nil, nil
	}
	if err := deps.SaveSettings(ctx, view); err != nil {
		return nil, err
	}
	return view, nil
}

// UpdateDependencies wires an update of the mihomo native settings.
type UpdateDependencies struct {
	ConfigDir    string
	LoadSettings SettingsLoader
	SaveSettings SettingsSaver
	ReplaceRefs  ResourceRefReplacer
	AfterApply   func()
}

// Update applies view onto the current mihomo native settings, persisting the
// resulting plan to both the settings store and the on-disk config before
// triggering the after-apply hook. It returns the freshly persisted view.
func Update(ctx context.Context, deps UpdateDependencies, view *proxygatewayv1.ProxyGatewayMihomoNativeConfig) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if deps.LoadSettings == nil {
		return nil, appcore.InternalError("mihomo native settings repository is required", nil)
	}
	current, err := loadUpdateCurrent(ctx, deps.LoadSettings)
	if err != nil {
		return nil, err
	}
	plan, err := mihomonative.BuildUpdatePlan(current, view)
	if err != nil {
		return nil, err
	}
	if err := persistUpdatePlan(ctx, deps, plan); err != nil {
		return nil, err
	}
	if deps.AfterApply != nil {
		deps.AfterApply()
	}
	return deps.LoadSettings(ctx)
}

func loadUpdateCurrent(ctx context.Context, load SettingsLoader) (mihomonative.ConfigFile, error) {
	currentView, err := load(ctx)
	if err != nil {
		return mihomonative.ConfigFile{}, appcore.InternalError("load mihomo native settings", err)
	}
	current, err := mihomonative.ConfigFromSettings(currentView)
	if err != nil {
		return mihomonative.ConfigFile{}, appcore.InternalError("load mihomo native settings", err)
	}
	return current, nil
}

func persistUpdatePlan(ctx context.Context, deps UpdateDependencies, plan mihomonative.UpdatePlan) error {
	if err := deps.SaveSettings(ctx, mihomonative.SettingsFromConfig(plan.Config)); err != nil {
		return appcore.InternalError("save mihomo native settings", err)
	}
	if err := mihomonative.SaveConfig(deps.ConfigDir, plan.Config); err != nil {
		return appcore.InternalError("save mihomo native config", err)
	}
	if _, err := deps.ReplaceRefs(ctx, plan.ResourceReplacements); err != nil {
		return appcore.InternalError("update mihomo native resource references", err)
	}
	return nil
}
