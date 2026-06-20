package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func updateMihomoNativeSettings(ctx context.Context, deps mihomoNativeUpdateDependencies, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if deps.Repository == nil {
		return nil, appcore.InternalError("mihomo native settings repository is required", nil)
	}
	current, err := loadMihomoNativeUpdateCurrent(ctx, deps.Repository)
	if err != nil {
		return nil, err
	}
	plan, err := mihomonative.BuildUpdatePlan(current, view)
	if err != nil {
		return nil, err
	}
	if err := persistMihomoNativeUpdatePlan(ctx, deps, plan); err != nil {
		return nil, err
	}
	runMihomoNativeUpdateAfterApply(deps.AfterApply)
	return deps.Repository.loadMihomoNative(ctx)
}
