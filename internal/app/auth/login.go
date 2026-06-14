package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type LoginRequest struct {
	Token      string
	Next       string
	FormSubmit bool
}

type RequestBodyReader func(*http.Request) ([]byte, error)

func ReadLoginRequest(req *http.Request, readBody RequestBodyReader) (LoginRequest, error) {
	body, err := readBody(req)
	if err != nil {
		return LoginRequest{}, err
	}
	return ParseLoginRequest(req, body)
}

type loginJSONRequest struct {
	Token string `json:"token"`
}

func ParseLoginRequest(req *http.Request, body []byte) (LoginRequest, error) {
	next := "/"
	if req != nil && req.URL != nil {
		next = SafeRedirect(req.URL.Query().Get("next"))
	}
	contentType := ""
	if req != nil {
		contentType = strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Type")))
	}
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return LoginRequest{Next: next, FormSubmit: true}, errors.New("invalid login form")
		}
		return LoginRequest{Token: strings.TrimSpace(values.Get("token")), Next: firstNonEmpty(values.Get("next"), next), FormSubmit: true}, nil
	}
	var payload loginJSONRequest
	if len(strings.TrimSpace(string(body))) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return LoginRequest{Next: next}, errors.New("invalid login request")
		}
	}
	return LoginRequest{Token: strings.TrimSpace(payload.Token), Next: next}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
