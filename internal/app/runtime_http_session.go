package app

import (
	"net/http"
	"time"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleAuthSession(ctx *gin.Context) {
	api.auth.WriteSessionResponse(ctx.Writer, api.sessionAuthenticated(ctx.Request))
}

func (api *runtimeHTTPAPI) handleAuthWebSocketToken(ctx *gin.Context) {
	if err := api.auth.WriteWebSocketTokenResponse(ctx.Writer, time.Now()); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) handleAuthLogin(ctx *gin.Context) {
	login, err := authapp.ReadLoginRequest(ctx.Request, readRequestBody)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	decision, err := api.auth.DecideLogin(login)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusUnauthorized)
		return
	}
	if err := api.auth.WriteLoginDecision(ctx.Writer, ctx.Request, decision, time.Now()); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) handleAuthLogout(ctx *gin.Context) {
	api.auth.WriteLogoutResponse(ctx.Writer, ctx.Request, ctx.Query("redirect"))
}

func (api *runtimeHTTPAPI) handleAuthLoginPage(ctx *gin.Context) {
	if err := api.auth.WriteLoginPageResponse(ctx.Writer, ctx.Request, api.sessionAuthenticated(ctx.Request)); err != nil {
		api.logger.Warn("render login page failed", "error_type", errorLogType(err))
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) redirectToLoginIfRequired(ctx *gin.Context) bool {
	return api.auth.WriteLoginRedirectIfRequired(ctx.Writer, ctx.Request, "/ui", time.Now())
}

func (api *runtimeHTTPAPI) sessionAuthenticated(req *http.Request) bool {
	return api.auth.SessionAuthenticated(req, time.Now())
}
