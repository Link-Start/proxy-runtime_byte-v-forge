package app

import (
	"net/http"
	"time"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleAuthSession(ctx *gin.Context) {
	api.writeAuthSession(ctx, api.sessionAuthenticated(ctx.Request))
}

func (api *runtimeHTTPAPI) handleAuthWebSocketToken(ctx *gin.Context) {
	token, err := api.auth.NewWebSocketToken(time.Now())
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	authapp.WriteWebSocketTokenResponse(ctx.Writer, token)
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
	if !decision.Authenticated {
		ctx.Redirect(http.StatusSeeOther, decision.RedirectURL)
		return
	}
	if err := api.auth.SetSessionCookie(ctx.Writer, ctx.Request, time.Now()); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	if decision.RedirectURL != "" {
		ctx.Redirect(http.StatusSeeOther, decision.RedirectURL)
		return
	}
	api.writeAuthSession(ctx, true)
}

func (api *runtimeHTTPAPI) handleAuthLogout(ctx *gin.Context) {
	api.auth.WriteLogoutResponse(ctx.Writer, ctx.Request, ctx.Query("redirect"))
}

func (api *runtimeHTTPAPI) handleAuthLoginPage(ctx *gin.Context) {
	page := authapp.LoginPageOptionsFromRequest(ctx.Request)
	if api.sessionAuthenticated(ctx.Request) {
		ctx.Redirect(http.StatusSeeOther, page.Next)
		return
	}
	if err := authapp.WriteLoginPage(ctx.Writer, page); err != nil {
		api.logger.Warn("render login page failed", "error", err)
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) writeAuthSession(ctx *gin.Context, authenticated bool) {
	authapp.WriteSessionResponse(ctx.Writer, authenticated, api.auth.Enabled())
}

func (api *runtimeHTTPAPI) redirectToLoginIfRequired(ctx *gin.Context) bool {
	redirectURL, required := api.auth.LoginRedirectIfRequired(ctx.Request, "/ui", time.Now())
	if !required {
		return false
	}
	ctx.Redirect(http.StatusSeeOther, redirectURL)
	return true
}

func (api *runtimeHTTPAPI) sessionAuthenticated(req *http.Request) bool {
	return api.auth.SessionAuthenticated(req, time.Now())
}
