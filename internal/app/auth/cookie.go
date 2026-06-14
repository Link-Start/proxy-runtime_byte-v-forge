package auth

import (
	"net/http"
	"time"
)

func NewSessionCookie(secret string, now time.Time, secure bool) (*http.Cookie, error) {
	value, err := SignSession(secret, now.Add(SessionTTL))
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(SessionTTL.Seconds()),
		Expires:  now.Add(SessionTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}, nil
}

func NewClearSessionCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}
