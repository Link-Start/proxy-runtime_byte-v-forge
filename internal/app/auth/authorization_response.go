package auth

import (
	"net/http"
	"time"
)

type UnauthorizedWriter func(http.ResponseWriter)

func WriteAuthorizationDecision(w http.ResponseWriter, req *http.Request, decision AuthorizationDecision, writeUnauthorized UnauthorizedWriter) bool {
	if decision.Authorized {
		return true
	}
	if decision.RedirectURL != "" {
		http.Redirect(w, req, decision.RedirectURL, http.StatusSeeOther)
		return false
	}
	if decision.Challenge != "" {
		w.Header().Set("WWW-Authenticate", decision.Challenge)
	}
	writeAuthorizationUnauthorized(w, writeUnauthorized)
	return false
}

func writeAuthorizationUnauthorized(w http.ResponseWriter, writeUnauthorized UnauthorizedWriter) {
	if writeUnauthorized != nil {
		writeUnauthorized(w)
		return
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (a Application) WriteLoginRedirectIfRequired(w http.ResponseWriter, req *http.Request, protectedPath string, now time.Time) bool {
	redirectURL, required := a.LoginRedirectIfRequired(req, protectedPath, now)
	if !required {
		return false
	}
	http.Redirect(w, req, redirectURL, http.StatusSeeOther)
	return true
}
