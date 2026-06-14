package auth

import (
	"net/http"
	"strings"
	"time"

	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
)

type Application struct {
	secret string
}

func NewApplication(secret string) Application {
	return Application{secret: strings.TrimSpace(secret)}
}

func (a Application) Enabled() bool {
	return a.secret != ""
}

func (a Application) Required(requestPath string) bool {
	return Required(a.secret, requestPath)
}

func (a Application) TokenMatches(token string) bool {
	return TokenMatches(token, a.secret)
}

func (a Application) NewSessionCookie(now time.Time, secure bool) (*http.Cookie, error) {
	return NewSessionCookie(a.secret, now, secure)
}

func (a Application) NewClearSessionCookie(secure bool) *http.Cookie {
	return NewClearSessionCookie(secure)
}

func (a Application) NewWebSocketToken(now time.Time) (string, error) {
	return NewWebSocketToken(a.secret, now)
}

func (a Application) SessionAuthenticated(req *http.Request, now time.Time) bool {
	if !a.Enabled() {
		return true
	}
	if req == nil {
		return false
	}
	cookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		return false
	}
	return VerifySession(cookie.Value, a.secret, now)
}

func (a Application) RequestAuthenticated(req *http.Request, now time.Time) bool {
	if a.SessionAuthenticated(req, now) {
		return true
	}
	if req == nil || req.URL == nil || !httpapi.PathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	return VerifySession(req.URL.Query().Get("session"), a.secret, now)
}

func (a Application) LoginRedirectPreferred(req *http.Request, controlPlanePrefix string) bool {
	if req == nil || req.URL == nil {
		return false
	}
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	if httpapi.PathInPrefix(req.URL.Path, controlPlanePrefix) || httpapi.PathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	accept := strings.ToLower(req.Header.Get("Accept"))
	return accept == "" || strings.Contains(accept, "text/html")
}
