package app

import (
	"bytes"
	"context"
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

	"github.com/gin-gonic/gin"
)

const mihomoDashboardEndpointID = "proxy-runtime-mihomo"

type mihomoProxyContextKey string

const mihomoProxyRequestPathKey mihomoProxyContextKey = "request_path"

func (api *runtimeHTTPAPI) registerMihomoDashboardRoutes(router *gin.Engine) {
	router.GET("/mihomo/dashboard", api.handleMihomoDashboard)
	router.GET("/mihomo/ui", redirectToTrailingSlash)
	router.Any("/mihomo/ui/*path", api.mihomoReverseProxy("/mihomo/ui/", "/ui/"))
	router.Any("/mihomo/controller", api.mihomoReverseProxy("/mihomo/controller", "/"))
	router.Any("/mihomo/controller/*path", api.mihomoReverseProxy("/mihomo/controller/", "/"))
}

func (api *runtimeHTTPAPI) handleMihomoDashboard(ctx *gin.Context) {
	api.writeMihomoDashboardBootstrap(ctx, "/mihomo/controller", "/mihomo/ui/#/overview")
}

func (api *runtimeHTTPAPI) writeMihomoDashboardBootstrap(ctx *gin.Context, endpointURL string, uiURL string) {
	payload, err := json.Marshal(struct {
		EndpointID   string `json:"endpointID"`
		EndpointURL  string `json:"endpointURL"`
		UIURL        string `json:"uiURL"`
		AuthRequired bool   `json:"authRequired"`
	}{
		EndpointID:   mihomoDashboardEndpointID,
		EndpointURL:  endpointURL,
		UIURL:        uiURL,
		AuthRequired: strings.TrimSpace(api.authToken) != "",
	})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Header("Cache-Control", "no-store")
	_, _ = fmt.Fprintf(ctx.Writer, `<!doctype html>
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
window.localStorage.setItem('proxyRuntimeControlAuthRequired', config.authRequired ? 'true' : 'false');
window.localStorage.setItem('endpointList', JSON.stringify([endpoint]));
window.localStorage.setItem('selectedEndpoint', endpoint.id);
window.location.replace(new URL(config.uiURL, window.location.origin).href);
</script>
</body>
</html>`, payload)
}

func (api *runtimeHTTPAPI) mihomoReverseProxy(mountPrefix string, upstreamPrefix string) gin.HandlerFunc {
	target, err := mihomoAPIURL(api.mihomoAPIAddr)
	if err != nil {
		return func(ctx *gin.Context) {
			writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		}
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(out *http.Request) {
		requestPath := out.URL.Path
		requestHost := out.Host
		*out = *out.WithContext(context.WithValue(out.Context(), mihomoProxyRequestPathKey, requestPath))
		out.URL.Scheme = target.Scheme
		out.URL.Host = target.Host
		out.URL.Path = joinMihomoProxyPath(upstreamPrefix, strings.TrimPrefix(requestPath, mountPrefix))
		out.Host = target.Host
		out.Header.Set("X-Forwarded-Host", requestHost)
		out.Header.Set("X-Forwarded-Proto", forwardedProto(out))
		api.forwardMihomoControllerAuthorization(out, requestPath)
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		writeHTTPError(w, err, http.StatusBadGateway)
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		applyMihomoDashboardCacheHeaders(resp, mihomoProxyRequestPath(resp.Request))
		return redactMihomoControllerErrorResponse(resp)
	}
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func mihomoProxyRequestPath(req *http.Request) string {
	if req != nil {
		if requestPath, ok := req.Context().Value(mihomoProxyRequestPathKey).(string); ok {
			return requestPath
		}
		if req.URL != nil {
			return req.URL.Path
		}
	}
	return ""
}

func applyMihomoDashboardCacheHeaders(resp *http.Response, requestPath string) {
	if resp == nil || !mihomoDashboardNoStorePath(requestPath, resp.Header.Get("Content-Type")) {
		return
	}
	resp.Header.Set("Cache-Control", "no-store")
	resp.Header.Set("Pragma", "no-cache")
	resp.Header.Set("Expires", "0")
}

func mihomoDashboardNoStorePath(requestPath string, contentType string) bool {
	cleanPath := strings.TrimSuffix(requestPath, "/")
	switch cleanPath {
	case "", "/mihomo/ui", "/mihomo/ui/index.html", "/sw.js", "/mihomo/ui/sw.js", "/config.js", "/mihomo/ui/config.js", "/manifest.webmanifest", "/mihomo/ui/manifest.webmanifest":
		return true
	}
	return strings.HasPrefix(strings.ToLower(contentType), "text/html")
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

func redirectToTrailingSlash(ctx *gin.Context) {
	target := *ctx.Request.URL
	target.Path += "/"
	ctx.Redirect(http.StatusTemporaryRedirect, target.String())
}
