package app

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"github.com/gin-gonic/gin"
)

type runtimeAuthSessionResponse struct {
	Authenticated bool `json:"authenticated"`
	AuthRequired  bool `json:"authRequired"`
}

type runtimeAuthLoginRequest struct {
	Token string `json:"token"`
}

type runtimeAuthWebSocketTokenResponse struct {
	Token string `json:"token"`
}

func (api *runtimeHTTPAPI) handleAuthSession(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")
	api.writeAuthSession(ctx, api.sessionAuthenticated(ctx.Request))
}

func (api *runtimeHTTPAPI) handleAuthWebSocketToken(ctx *gin.Context) {
	token, err := authapp.SignSession(api.authToken, time.Now().Add(authapp.WebSocketTokenTTL))
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Content-Type", "application/json")
	_ = json.NewEncoder(ctx.Writer).Encode(runtimeAuthWebSocketTokenResponse{Token: token})
}

func (api *runtimeHTTPAPI) handleAuthLogin(ctx *gin.Context) {
	token, next, formSubmit, err := readRuntimeLoginRequest(ctx.Request)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	if !authapp.TokenMatches(token, api.authToken) {
		if formSubmit {
			ctx.Redirect(http.StatusSeeOther, authapp.LoginRedirectWithError(next))
			return
		}
		writeHTTPError(ctx.Writer, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}
	if err := api.setSessionCookie(ctx); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	if formSubmit {
		ctx.Redirect(http.StatusSeeOther, authapp.SafeRedirect(next))
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
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Header("Cache-Control", "no-store")
	errorText := ""
	if strings.TrimSpace(ctx.Query("error")) != "" {
		errorText = "密钥无效"
	}
	data := struct {
		Next      string
		ErrorText string
	}{
		Next:      next,
		ErrorText: errorText,
	}
	if err := runtimeLoginTemplate.Execute(ctx.Writer, data); err != nil {
		api.logger.Warn("render login page failed", "error", err)
	}
}

func (api *runtimeHTTPAPI) writeAuthSession(ctx *gin.Context, authenticated bool) {
	ctx.Header("Content-Type", "application/json")
	_ = json.NewEncoder(ctx.Writer).Encode(runtimeAuthSessionResponse{
		Authenticated: authenticated,
		AuthRequired:  strings.TrimSpace(api.authToken) != "",
	})
}

func (api *runtimeHTTPAPI) redirectToLoginIfRequired(ctx *gin.Context) bool {
	if !api.authRequired("/ui") || api.sessionAuthenticated(ctx.Request) {
		return false
	}
	api.redirectToLogin(ctx, ctx.Request.URL.RequestURI())
	return true
}

func (api *runtimeHTTPAPI) redirectLoginPreferred(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	if httpapi.PathInPrefix(req.URL.Path, controlPlaneHTTPPrefix) || httpapi.PathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	accept := strings.ToLower(req.Header.Get("Accept"))
	return accept == "" || strings.Contains(accept, "text/html")
}

func (api *runtimeHTTPAPI) redirectToLogin(ctx *gin.Context, next string) {
	loginURL := "/login"
	if safe := authapp.SafeRedirect(next); safe != "/" {
		loginURL += "?next=" + url.QueryEscape(safe)
	}
	ctx.Redirect(http.StatusSeeOther, loginURL)
}

func (api *runtimeHTTPAPI) sessionAuthenticated(req *http.Request) bool {
	if strings.TrimSpace(api.authToken) == "" {
		return true
	}
	cookie, err := req.Cookie(authapp.SessionCookieName)
	if err != nil {
		return false
	}
	return authapp.VerifySession(cookie.Value, api.authToken, time.Now())
}

func (api *runtimeHTTPAPI) requestAuthenticated(req *http.Request) bool {
	if api.sessionAuthenticated(req) {
		return true
	}
	if req == nil || req.URL == nil || !httpapi.PathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	return authapp.VerifySession(req.URL.Query().Get("session"), api.authToken, time.Now())
}

func (api *runtimeHTTPAPI) setSessionCookie(ctx *gin.Context) error {
	value, err := authapp.SignSession(api.authToken, time.Now().Add(authapp.SessionTTL))
	if err != nil {
		return err
	}
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     authapp.SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(authapp.SessionTTL.Seconds()),
		Expires:  time.Now().Add(authapp.SessionTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   httpapi.ForwardedProto(ctx.Request) == "https",
	})
	return nil
}

func (api *runtimeHTTPAPI) clearSessionCookie(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     authapp.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   httpapi.ForwardedProto(ctx.Request) == "https",
	})
}

var runtimeLoginTemplate = template.Must(template.New("runtime-login").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Proxy Runtime 登录</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: radial-gradient(circle at top, #1f2a44 0, #0b1020 45%, #05070d 100%); color: #e5e7eb; }
    main { width: min(92vw, 420px); border: 1px solid rgba(148, 163, 184, .2); border-radius: 24px; padding: 28px; background: rgba(15, 23, 42, .82); box-shadow: 0 24px 80px rgba(0, 0, 0, .35); backdrop-filter: blur(18px); }
    h1 { margin: 0 0 8px; font-size: 24px; }
    p { margin: 0 0 22px; color: #94a3b8; }
    label { display: block; margin-bottom: 8px; color: #cbd5e1; font-size: 14px; }
    input { box-sizing: border-box; width: 100%; border: 1px solid rgba(148, 163, 184, .25); border-radius: 14px; padding: 13px 14px; background: rgba(2, 6, 23, .7); color: #f8fafc; outline: none; }
    input:focus { border-color: #38bdf8; box-shadow: 0 0 0 3px rgba(56, 189, 248, .18); }
    button { width: 100%; margin-top: 18px; border: 0; border-radius: 14px; padding: 13px 16px; color: #031018; background: linear-gradient(135deg, #38bdf8, #22c55e); font-weight: 700; cursor: pointer; }
    .error { margin-bottom: 14px; border: 1px solid rgba(248, 113, 113, .28); border-radius: 12px; padding: 10px 12px; color: #fecaca; background: rgba(127, 29, 29, .25); }
  </style>
</head>
<body>
  <main>
    <h1>Proxy Runtime</h1>
    <p>输入控制面密钥后继续访问 MetaCubeXD。</p>
    {{if .ErrorText}}<div class="error">{{.ErrorText}}</div>{{end}}
    <form method="post" action="/api/auth/login?next={{urlquery .Next}}">
      <input type="hidden" name="next" value="{{.Next}}">
      <label for="token">控制面密钥</label>
      <input id="token" name="token" type="password" autocomplete="current-password" autofocus required>
      <button type="submit">登录</button>
    </form>
  </main>
</body>
</html>`))
