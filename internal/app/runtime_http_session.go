package app

import (
	"errors"
	"net/http"
	"strings"
	"time"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
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
	login, err := readRuntimeLoginRequest(ctx.Request)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	if !api.auth.TokenMatches(login.Token) {
		if login.FormSubmit {
			ctx.Redirect(http.StatusSeeOther, authapp.LoginRedirectWithError(login.Next))
			return
		}
		writeHTTPError(ctx.Writer, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}
	if err := api.setSessionCookie(ctx); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	if login.FormSubmit {
		ctx.Redirect(http.StatusSeeOther, authapp.SafeRedirect(login.Next))
		return
	}
	api.writeAuthSession(ctx, true)
}

func (api *runtimeHTTPAPI) handleAuthLogout(ctx *gin.Context) {
	api.clearSessionCookie(ctx)
	if redirect := strings.TrimSpace(ctx.Query("redirect")); redirect != "" {
		ctx.Redirect(http.StatusSeeOther, authapp.SafeRedirect(redirect))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleAuthLoginPage(ctx *gin.Context) {
	next := authapp.SafeRedirect(ctx.Query("next"))
	if api.sessionAuthenticated(ctx.Request) {
		ctx.Redirect(http.StatusSeeOther, next)
		return
	}
	if err := authapp.WriteLoginPage(ctx.Writer, authapp.LoginPageOptions{Next: next, ShowError: strings.TrimSpace(ctx.Query("error")) != ""}); err != nil {
		api.logger.Warn("render login page failed", "error", err)
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) writeAuthSession(ctx *gin.Context, authenticated bool) {
	authapp.WriteSessionResponse(ctx.Writer, authenticated, api.auth.Enabled())
}

func (api *runtimeHTTPAPI) redirectToLoginIfRequired(ctx *gin.Context) bool {
	if !api.authRequired("/ui") || api.sessionAuthenticated(ctx.Request) {
		return false
	}
	api.redirectToLogin(ctx, ctx.Request.URL.RequestURI())
	return true
}

func (api *runtimeHTTPAPI) redirectLoginPreferred(req *http.Request) bool {
	return api.auth.LoginRedirectPreferred(req, controlPlaneHTTPPrefix)
}

func (api *runtimeHTTPAPI) redirectToLogin(ctx *gin.Context, next string) {
	ctx.Redirect(http.StatusSeeOther, authapp.LoginRedirect(next))
}

func (api *runtimeHTTPAPI) sessionAuthenticated(req *http.Request) bool {
	return api.auth.SessionAuthenticated(req, time.Now())
}

func (api *runtimeHTTPAPI) requestAuthenticated(req *http.Request) bool {
	return api.auth.RequestAuthenticated(req, time.Now())
}

func (api *runtimeHTTPAPI) setSessionCookie(ctx *gin.Context) error {
	cookie, err := api.auth.NewSessionCookie(time.Now(), httpapi.ForwardedProto(ctx.Request) == "https")
	if err != nil {
		return err
	}
	http.SetCookie(ctx.Writer, cookie)
	return nil
}

func (api *runtimeHTTPAPI) clearSessionCookie(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, api.auth.NewClearSessionCookie(httpapi.ForwardedProto(ctx.Request) == "https"))
}

func readRuntimeLoginRequest(req *http.Request) (authapp.LoginRequest, error) {
	body, err := readRequestBody(req)
	if err != nil {
		return authapp.LoginRequest{}, err
	}
	return authapp.ParseLoginRequest(req, body)
}
