package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/random"
	"github.com/gin-gonic/gin"
)

const (
	runtimeSessionCookieName = "proxy_runtime_session"
	runtimeSessionVersion    = "v1"
	runtimeSessionTTL        = 12 * time.Hour
)

type runtimeAuthSessionResponse struct {
	Authenticated bool `json:"authenticated"`
	AuthRequired  bool `json:"authRequired"`
}

type runtimeAuthLoginRequest struct {
	Token string `json:"token"`
}

func (api *runtimeHTTPAPI) handleAuthSession(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")
	api.writeAuthSession(ctx, api.sessionAuthenticated(ctx.Request))
}

func (api *runtimeHTTPAPI) handleAuthLogin(ctx *gin.Context) {
	token, next, formSubmit, err := readRuntimeLoginRequest(ctx.Request)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	if !authTokenMatches(token, api.authToken) {
		if formSubmit {
			ctx.Redirect(http.StatusSeeOther, loginRedirectWithError(next))
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
		ctx.Redirect(http.StatusSeeOther, safeAuthRedirect(next))
		return
	}
	api.writeAuthSession(ctx, true)
}

func (api *runtimeHTTPAPI) handleAuthLogout(ctx *gin.Context) {
	api.clearSessionCookie(ctx)
	if redirect := strings.TrimSpace(ctx.Query("redirect")); redirect != "" {
		ctx.Redirect(http.StatusSeeOther, safeAuthRedirect(redirect))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleAuthLoginPage(ctx *gin.Context) {
	next := safeAuthRedirect(ctx.Query("next"))
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
	if pathInPrefix(req.URL.Path, controlPlaneHTTPPrefix) || pathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	accept := strings.ToLower(req.Header.Get("Accept"))
	return accept == "" || strings.Contains(accept, "text/html")
}

func (api *runtimeHTTPAPI) redirectToLogin(ctx *gin.Context, next string) {
	loginURL := "/login"
	if safe := safeAuthRedirect(next); safe != "/" {
		loginURL += "?next=" + url.QueryEscape(safe)
	}
	ctx.Redirect(http.StatusSeeOther, loginURL)
}

func (api *runtimeHTTPAPI) sessionAuthenticated(req *http.Request) bool {
	if strings.TrimSpace(api.authToken) == "" {
		return true
	}
	cookie, err := req.Cookie(runtimeSessionCookieName)
	if err != nil {
		return false
	}
	return verifyRuntimeSession(cookie.Value, api.authToken, time.Now())
}

func (api *runtimeHTTPAPI) setSessionCookie(ctx *gin.Context) error {
	value, err := signRuntimeSession(api.authToken, time.Now().Add(runtimeSessionTTL))
	if err != nil {
		return err
	}
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     runtimeSessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(runtimeSessionTTL.Seconds()),
		Expires:  time.Now().Add(runtimeSessionTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   forwardedProto(ctx.Request) == "https",
	})
	return nil
}

func (api *runtimeHTTPAPI) clearSessionCookie(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     runtimeSessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   forwardedProto(ctx.Request) == "https",
	})
}

func signRuntimeSession(secret string, expiresAt time.Time) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("auth token is empty")
	}
	nonce, err := random.Hex(16)
	if err != nil {
		return "", err
	}
	message := strings.Join([]string{runtimeSessionVersion, strconv.FormatInt(expiresAt.Unix(), 10), nonce}, ".")
	return message + "." + runtimeSessionSignature(secret, message), nil
}

func verifyRuntimeSession(value string, secret string, now time.Time) bool {
	secret = strings.TrimSpace(secret)
	parts := strings.Split(strings.TrimSpace(value), ".")
	if secret == "" || len(parts) != 4 || parts[0] != runtimeSessionVersion {
		return false
	}
	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.Unix() > expiresAt {
		return false
	}
	message := strings.Join(parts[:3], ".")
	expected := runtimeSessionSignature(secret, message)
	return subtle.ConstantTimeCompare([]byte(parts[3]), []byte(expected)) == 1
}

func runtimeSessionSignature(secret string, message string) string {
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(secret)))
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func readRuntimeLoginRequest(req *http.Request) (token string, next string, formSubmit bool, err error) {
	next = safeAuthRedirect(req.URL.Query().Get("next"))
	contentType := strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Type")))
	body, err := readRequestBody(req)
	if err != nil {
		return "", next, false, err
	}
	switch {
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		values, parseErr := url.ParseQuery(string(body))
		if parseErr != nil {
			return "", next, true, invalidArgument("invalid login form", parseErr)
		}
		return strings.TrimSpace(values.Get("token")), firstNonEmpty(values.Get("next"), next), true, nil
	default:
		var payload runtimeAuthLoginRequest
		if len(strings.TrimSpace(string(body))) > 0 {
			if parseErr := json.Unmarshal(body, &payload); parseErr != nil {
				return "", next, false, invalidArgument("invalid login request", parseErr)
			}
		}
		return strings.TrimSpace(payload.Token), next, false, nil
	}
}

func loginRedirectWithError(next string) string {
	target := "/login?error=1"
	if safe := safeAuthRedirect(next); safe != "/" {
		target += "&next=" + url.QueryEscape(safe)
	}
	return target
}

func safeAuthRedirect(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/"
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "/"
	}
	return parsed.String()
}

func mihomoControllerUpstreamRawQuery(rawQuery string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}
	values.Del("token")
	return values.Encode()
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
