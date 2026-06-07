package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const mihomoDashboardEndpointID = "proxy-runtime-mihomo"

func (api *runtimeHTTPAPI) registerMihomoDashboardRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/mihomo/dashboard", api.handleMihomoDashboard)
	mux.HandleFunc(prefix+"/mihomo/ui", redirectToTrailingSlash)
	mux.Handle(prefix+"/mihomo/ui/", api.mihomoReverseProxy(prefix+"/mihomo/ui/", "/ui/"))
	mux.Handle(prefix+"/mihomo/controller", api.mihomoReverseProxy(prefix+"/mihomo/controller", "/"))
	mux.Handle(prefix+"/mihomo/controller/", api.mihomoReverseProxy(prefix+"/mihomo/controller/", "/"))
}

func (api *runtimeHTTPAPI) handleMihomoDashboard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	basePath := strings.TrimSuffix(req.URL.Path, "/mihomo/dashboard")
	payload, err := json.Marshal(struct {
		EndpointID  string `json:"endpointID"`
		EndpointURL string `json:"endpointURL"`
		UIURL       string `json:"uiURL"`
	}{
		EndpointID:  mihomoDashboardEndpointID,
		EndpointURL: basePath + "/mihomo/controller",
		UIURL:       basePath + "/mihomo/ui/#/proxies",
	})
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = fmt.Fprintf(w, `<!doctype html>
<html>
<head><meta charset="utf-8"><title>Proxy Runtime</title></head>
<body>
<script>
const config = %s;
const endpoint = {
  id: config.endpointID,
  url: new URL(config.endpointURL, window.location.origin).href.replace(/\/$/, ''),
  secret: ''
};
let endpoints = [];
try {
  const parsed = JSON.parse(window.localStorage.getItem('endpointList') || '[]');
  if (Array.isArray(parsed)) {
    endpoints = parsed.filter((item) => item && item.id !== endpoint.id && item.url !== endpoint.url);
  }
} catch (_) {}
window.localStorage.setItem('endpointList', JSON.stringify([endpoint, ...endpoints]));
window.localStorage.setItem('selectedEndpoint', endpoint.id);
window.location.replace(new URL(config.uiURL, window.location.origin).href);
</script>
</body>
</html>`, payload)
}

func (api *runtimeHTTPAPI) mihomoReverseProxy(mountPrefix string, upstreamPrefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		target, err := mihomoAPIURL(api.mihomoAPIAddr)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadGateway)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Director = func(out *http.Request) {
			out.URL.Scheme = target.Scheme
			out.URL.Host = target.Host
			out.URL.Path = joinMihomoProxyPath(upstreamPrefix, strings.TrimPrefix(req.URL.Path, mountPrefix))
			out.Host = target.Host
			out.Header.Set("X-Forwarded-Host", req.Host)
			out.Header.Set("X-Forwarded-Proto", forwardedProto(req))
		}
		proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
			writeHTTPError(w, err, http.StatusBadGateway)
		}
		proxy.ModifyResponse = redactMihomoControllerErrorResponse
		proxy.ServeHTTP(w, req)
	})
}

func redactMihomoControllerErrorResponse(resp *http.Response) error {
	if resp == nil || resp.StatusCode < http.StatusBadRequest || resp.Body == nil {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	data = redactURLCredentials(data)
	resp.Body = io.NopCloser(bytes.NewReader(data))
	resp.ContentLength = int64(len(data))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
	return nil
}

var sensitiveURLPattern = regexp.MustCompile(`https?://[^\s"'<>)\\]+`)

func redactURLCredentials(data []byte) []byte {
	return sensitiveURLPattern.ReplaceAll(data, []byte("[redacted-url]"))
}

func mihomoAPIURL(addr string) (*url.URL, error) {
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

func joinMihomoProxyPath(prefix string, suffix string) string {
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

func forwardedProto(req *http.Request) string {
	if value := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto")); value != "" {
		return value
	}
	if req.TLS != nil {
		return "https"
	}
	return "http"
}

func redirectToTrailingSlash(w http.ResponseWriter, req *http.Request) {
	target := *req.URL
	target.Path += "/"
	http.Redirect(w, req, target.String(), http.StatusTemporaryRedirect)
}
