package auth

import (
	"net/http"
	"strings"
	"time"

	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
)

type Application struct {
	controlSecret string
	serviceSecret string
}

func NewApplication(controlSecret string, serviceSecret string) Application {
	return Application{
		controlSecret: strings.TrimSpace(controlSecret),
		serviceSecret: strings.TrimSpace(serviceSecret),
	}
}

func (a Application) Enabled() bool {
	return a.controlSecret != ""
}

func (a Application) Required(requestPath string) bool {
	return Required(a.controlSecret, requestPath)
}

func (a Application) RequestRequired(method string, requestPath string) bool {
	if a.serviceSecret != "" && ServiceRuntimePath(method, requestPath) {
		return true
	}
	return a.Required(requestPath)
}

func (a Application) TokenMatches(token string) bool {
	return TokenMatches(token, a.controlSecret)
}

func (a Application) NewSessionCookie(now time.Time, secure bool) (*http.Cookie, error) {
	return NewSessionCookie(a.controlSecret, now, secure)
}

func (a Application) NewClearSessionCookie(secure bool) *http.Cookie {
	return NewClearSessionCookie(secure)
}

func (a Application) SetSessionCookie(w http.ResponseWriter, req *http.Request, now time.Time) error {
	cookie, err := a.NewSessionCookie(now, httpapi.ForwardedProto(req) == "https")
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

func (a Application) ClearSessionCookie(w http.ResponseWriter, req *http.Request) {
	http.SetCookie(w, a.NewClearSessionCookie(httpapi.ForwardedProto(req) == "https"))
}

func (a Application) WriteLogoutResponse(w http.ResponseWriter, req *http.Request, redirect string) {
	a.ClearSessionCookie(w, req)
	if target := LogoutRedirect(redirect); target != "" {
		http.Redirect(w, req, target, http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a Application) WriteLoginDecision(w http.ResponseWriter, req *http.Request, decision LoginDecision, now time.Time) error {
	if !decision.Authenticated {
		http.Redirect(w, req, decision.RedirectURL, http.StatusSeeOther)
		return nil
	}
	if err := a.SetSessionCookie(w, req, now); err != nil {
		return err
	}
	if decision.RedirectURL != "" {
		http.Redirect(w, req, decision.RedirectURL, http.StatusSeeOther)
		return nil
	}
	a.WriteSessionResponse(w, true)
	return nil
}

func (a Application) WriteLoginPageResponse(w http.ResponseWriter, req *http.Request, authenticated bool) error {
	page := LoginPageOptionsFromRequest(req)
	if authenticated {
		http.Redirect(w, req, page.Next, http.StatusSeeOther)
		return nil
	}
	return WriteLoginPage(w, page)
}

func (a Application) WriteSessionResponse(w http.ResponseWriter, authenticated bool) {
	WriteSessionResponse(w, authenticated, a.Enabled())
}

func (a Application) WriteWebSocketTokenResponse(w http.ResponseWriter, now time.Time) error {
	token, err := a.NewWebSocketToken(now)
	if err != nil {
		return err
	}
	WriteWebSocketTokenResponse(w, token)
	return nil
}

func (a Application) NewWebSocketToken(now time.Time) (string, error) {
	return NewWebSocketToken(a.controlSecret, now)
}

func (a Application) SessionAuthenticated(req *http.Request, now time.Time) bool {
	if !a.Enabled() {
		return true
	}
	return a.controlSessionAuthenticated(req, now)
}

func (a Application) controlSessionAuthenticated(req *http.Request, now time.Time) bool {
	if !a.Enabled() {
		return false
	}
	if req == nil {
		return false
	}
	cookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		return false
	}
	return VerifySession(cookie.Value, a.controlSecret, now)
}

func (a Application) RequestAuthenticated(req *http.Request, now time.Time) bool {
	if a.controlSessionAuthenticated(req, now) || a.serviceRequestAuthenticated(req) {
		return true
	}
	if req == nil || req.URL == nil || !httpapi.PathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	return VerifySession(req.URL.Query().Get("session"), a.controlSecret, now)
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
