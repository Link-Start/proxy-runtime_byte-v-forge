package dashboard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

func ControllerUpstreamRawQuery(rawQuery string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}
	values.Del("token")
	values.Del("session")
	return values.Encode()
}

func ApplyCacheHeaders(resp *http.Response, requestPath string) {
	if resp == nil || !dashboardNoStorePath(requestPath, resp.Header.Get("Content-Type")) {
		return
	}
	resp.Header.Set("Cache-Control", "no-store")
	resp.Header.Set("Pragma", "no-cache")
	resp.Header.Set("Expires", "0")
}

func RedactControllerErrorResponse(resp *http.Response) error {
	if resp == nil || resp.StatusCode < http.StatusBadRequest || resp.Body == nil {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	data = RedactSensitiveControllerData(data)
	resp.Body = io.NopCloser(bytes.NewReader(data))
	resp.ContentLength = int64(len(data))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
	return nil
}

func RedactSensitiveControllerData(data []byte) []byte {
	for _, pattern := range sensitiveResponsePatterns {
		data = pattern.expr.ReplaceAll(data, pattern.replacement)
	}
	return data
}

func APIURL(addr string) (*url.URL, error) {
	raw := strings.TrimSpace(addr)
	if raw == "" {
		return nil, errors.New("mihomo api address is empty")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Host == "" {
		return nil, errors.New("mihomo api address is invalid")
	}
	return parsed, nil
}

func JoinProxyPath(prefix string, suffix string) string {
	prefix = "/" + strings.Trim(strings.TrimSpace(prefix), "/")
	suffix = strings.TrimLeft(suffix, "/")
	if suffix == "" {
		if strings.HasSuffix(prefix, "/") {
			return prefix
		}
		return prefix + "/"
	}
	joined := path.Join(prefix, suffix)
	if strings.HasSuffix(suffix, "/") && !strings.HasSuffix(joined, "/") {
		return joined + "/"
	}
	return joined
}

func dashboardNoStorePath(requestPath string, contentType string) bool {
	cleanPath := strings.TrimSuffix(requestPath, "/")
	switch cleanPath {
	case "", "/mihomo/ui", "/mihomo/ui/index.html", "/sw.js", "/mihomo/ui/sw.js", "/config.js", "/mihomo/ui/config.js", "/manifest.webmanifest", "/mihomo/ui/manifest.webmanifest":
		return true
	}
	return strings.HasPrefix(strings.ToLower(contentType), "text/html")
}

type sensitiveResponsePattern struct {
	expr        *regexp.Regexp
	replacement []byte
}

var sensitiveResponsePatterns = []sensitiveResponsePattern{
	{expr: regexp.MustCompile(`https?://[^\s"'<>)\\]+`), replacement: []byte("[redacted-url]")},
	{expr: regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)[^\s"'<>)\\]+`), replacement: []byte("${1}[redacted]")},
	{expr: regexp.MustCompile(`(?i)(["']?(?:token|secret|password|session)["']?\s*[:=]\s*["']?)[^"',\s&}]+`), replacement: []byte("${1}[redacted]")},
}
