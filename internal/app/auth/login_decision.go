package auth

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

type LoginDecision struct {
	Authenticated bool
	RedirectURL   string
}

func (a Application) DecideLogin(login LoginRequest) (LoginDecision, error) {
	if !a.TokenMatches(login.Token) {
		if login.FormSubmit {
			return LoginDecision{RedirectURL: LoginRedirectWithError(login.Next)}, nil
		}
		return LoginDecision{}, ErrUnauthorized
	}
	if login.FormSubmit {
		return LoginDecision{Authenticated: true, RedirectURL: SafeRedirect(login.Next)}, nil
	}
	return LoginDecision{Authenticated: true}, nil
}
