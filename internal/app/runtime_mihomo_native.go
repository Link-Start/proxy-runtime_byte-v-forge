package app

import (
	"context"
	"errors"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	mihomoapp "github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative/application"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings/application"
	"github.com/gin-gonic/gin"
)

type runtimeMihomoNativeApplyScheduler struct {
	exitCheckCache   *proxycheck.ExitCheckCache
	markApplyPending func()
	requestReconcile func()
}

func newRuntimeMihomoNativeApplyScheduler(runtime *Runtime) runtimeMihomoNativeApplyScheduler {
	if runtime == nil {
		return runtimeMihomoNativeApplyScheduler{}
	}
	return runtimeMihomoNativeApplyScheduler{
		exitCheckCache:   runtime.exitCheckCache,
		markApplyPending: runtime.markSettingsApplyPending,
		requestReconcile: runtime.requestReconcile,
	}
}

func (s runtimeMihomoNativeApplyScheduler) Schedule() {
	if s.exitCheckCache != nil {
		s.exitCheckCache.Clear()
	}
	if s.markApplyPending != nil {
		s.markApplyPending()
	}
	if s.requestReconcile != nil {
		s.requestReconcile()
	}
}

func (api *runtimeHTTPAPI) handleMihomoNativeConfig(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		api.handleGetMihomoNativeConfig(ctx)
	case http.MethodPost, http.MethodPut:
		api.handleUpdateMihomoNativeConfig(ctx)
	}
}

func (api *runtimeHTTPAPI) handleGetMihomoNativeConfig(ctx *gin.Context) {
	response, err := api.settings.GetMihomoNative(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest{})
	if err != nil {
		writeSettingsLoadHTTPError(ctx, appcore.InternalError("load mihomo native config", err))
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleUpdateMihomoNativeConfig(ctx *gin.Context) {
	var req proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest
	if !api.readProto(ctx, &req) {
		return
	}
	response, err := api.settings.UpdateMihomoNative(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, settingsapp.ErrMihomoNativeUpdateUnavailable) {
			err = appcore.InternalError("mihomo native settings update unavailable", err)
		}
		writeSettingsUpdateHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}

func (r *Runtime) projectMihomoNativeSettings(ctx context.Context) error {
	if r == nil || r.settings == nil {
		return nil
	}
	return mihomoapp.Project(ctx, mihomoapp.ProjectionDependencies{
		ConfigDir:    r.cfg.Mihomo.ConfigDir,
		LoadSettings: r.settings.LoadMihomoNative,
		SaveSettings: r.settings.SaveMihomoNative,
	})
}
