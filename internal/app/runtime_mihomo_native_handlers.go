package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleMihomoNativeConfig(ctx *gin.Context) {
	runtime := api.service.settings.runtime
	switch ctx.Request.Method {
	case http.MethodGet:
		settings, err := mihomoNativeSettings(runtime)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
			return
		}
		writeJSON(ctx, settings)
	case http.MethodPost, http.MethodPut:
		var settings mihomoNativeSettingsView
		if !readJSON(ctx, &settings) {
			return
		}
		next, err := updateMihomoNativeSettings(ctx.Request.Context(), runtime, settings)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		writeJSON(ctx, next)
	}
}

func readJSON(ctx *gin.Context, dst any) bool {
	defer ctx.Request.Body.Close()
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 1<<20))
	if err != nil {
		writeHTTPError(ctx.Writer, invalidArgument("read request body", err), http.StatusBadRequest)
		return false
	}
	if len(body) == 0 {
		writeHTTPError(ctx.Writer, invalidArgument("request body is required", nil), http.StatusBadRequest)
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		writeHTTPError(ctx.Writer, invalidArgument("parse request body", err), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(ctx *gin.Context, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		writeHTTPError(ctx.Writer, internalError("marshal response", err), http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Type", "application/json")
	_, _ = ctx.Writer.Write(data)
}
