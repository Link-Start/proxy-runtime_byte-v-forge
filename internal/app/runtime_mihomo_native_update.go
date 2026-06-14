package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type mihomoNativeUpdateRepository interface {
	loadMihomoNative(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	saveMihomoNative(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error
	replaceMihomoResourceRefs(context.Context, map[string]mihomoNativeResourceReplacement) (bool, error)
}

type mihomoNativeUpdateDependencies struct {
	Repository mihomoNativeUpdateRepository
	ConfigDir  string
	AfterApply func()
}

func updateMihomoNativeSettings(ctx context.Context, deps mihomoNativeUpdateDependencies, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if deps.Repository == nil {
		return nil, internalError("mihomo native settings repository is required", nil)
	}
	current, err := loadMihomoNativeUpdateCurrent(ctx, deps.Repository)
	if err != nil {
		return nil, err
	}
	plan, err := buildMihomoNativeUpdatePlan(current, view)
	if err != nil {
		return nil, err
	}
	if err := persistMihomoNativeUpdatePlan(ctx, deps, plan); err != nil {
		return nil, err
	}
	if deps.AfterApply != nil {
		deps.AfterApply()
	}
	return deps.Repository.loadMihomoNative(ctx)
}
