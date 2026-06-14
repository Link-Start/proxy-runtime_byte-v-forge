package auth

import (
	"net/http"
	"time"
)

const cookieAuthChallenge = "Cookie"

type AuthorizationDecision struct {
	Authorized  bool
	RedirectURL string
	Challenge   string
}

func (a Application) Authorize(req *http.Request, now time.Time, controlPlanePrefix string) AuthorizationDecision {
	requestPath := ""
	if req != nil && req.URL != nil {
		requestPath = req.URL.Path
	}
	if !a.Required(requestPath) || !a.Enabled() || a.RequestAuthenticated(req, now) {
		return AuthorizationDecision{Authorized: true}
	}
	if a.LoginRedirectPreferred(req, controlPlanePrefix) {
		return AuthorizationDecision{RedirectURL: LoginRedirect(requestURI(req))}
	}
	return AuthorizationDecision{Challenge: cookieAuthChallenge}
}

func (a Application) LoginRedirectIfRequired(req *http.Request, protectedPath string, now time.Time) (string, bool) {
	if !a.Required(protectedPath) || a.SessionAuthenticated(req, now) {
		return "", false
	}
	return LoginRedirect(requestURI(req)), true
}

func requestURI(req *http.Request) string {
	if req == nil || req.URL == nil {
		return "/"
	}
	return req.URL.RequestURI()
}
