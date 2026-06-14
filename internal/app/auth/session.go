package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/random"
)

const (
	SessionCookieName = "proxy_runtime_session"
	SessionTTL        = 12 * time.Hour
	sessionVersion    = "v1"
)

func TokenMatches(actual string, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func SignSession(secret string, expiresAt time.Time) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("auth token is empty")
	}
	nonce, err := random.Hex(16)
	if err != nil {
		return "", err
	}
	message := strings.Join([]string{sessionVersion, strconv.FormatInt(expiresAt.Unix(), 10), nonce}, ".")
	return message + "." + sessionSignature(secret, message), nil
}

func VerifySession(value string, secret string, now time.Time) bool {
	secret = strings.TrimSpace(secret)
	parts := strings.Split(strings.TrimSpace(value), ".")
	if secret == "" || len(parts) != 4 || parts[0] != sessionVersion {
		return false
	}
	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.Unix() > expiresAt {
		return false
	}
	message := strings.Join(parts[:3], ".")
	expected := sessionSignature(secret, message)
	return subtle.ConstantTimeCompare([]byte(parts[3]), []byte(expected)) == 1
}

func LoginRedirect(next string) string {
	target := "/login"
	if safe := SafeRedirect(next); safe != "/" {
		target += "?next=" + url.QueryEscape(safe)
	}
	return target
}

func LoginRedirectWithError(next string) string {
	target := "/login?error=1"
	if safe := SafeRedirect(next); safe != "/" {
		target += "&next=" + url.QueryEscape(safe)
	}
	return target
}

func SafeRedirect(value string) string {
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

func sessionSignature(secret string, message string) string {
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(secret)))
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
