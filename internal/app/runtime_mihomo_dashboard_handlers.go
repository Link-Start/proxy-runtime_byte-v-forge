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

	"github.com/gin-gonic/gin"
)

const mihomoDashboardEndpointID = "proxy-runtime-mihomo"

func (api *runtimeHTTPAPI) registerMihomoDashboardRoutes(router *gin.Engine) {
	router.GET("/mihomo/dashboard", api.handleMihomoDashboard)
	router.GET("/mihomo/ui", redirectToTrailingSlash)
	router.Any("/mihomo/ui/*path", api.mihomoReverseProxy("/mihomo/ui/", "/ui/"))
	router.Any("/mihomo/controller", api.mihomoReverseProxy("/mihomo/controller", "/"))
	router.Any("/mihomo/controller/*path", api.mihomoReverseProxy("/mihomo/controller/", "/"))
}

func (api *runtimeHTTPAPI) handleMihomoDashboard(ctx *gin.Context) {
	api.writeMihomoDashboardBootstrap(ctx, "/mihomo/controller", "/mihomo/ui/#/proxies")
}

func (api *runtimeHTTPAPI) writeMihomoDashboardBootstrap(ctx *gin.Context, endpointURL string, uiURL string) {
	payload, err := json.Marshal(struct {
		EndpointID  string `json:"endpointID"`
		EndpointURL string `json:"endpointURL"`
		UIURL       string `json:"uiURL"`
	}{
		EndpointID:  mihomoDashboardEndpointID,
		EndpointURL: endpointURL,
		UIURL:       uiURL,
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

func (api *runtimeHTTPAPI) mihomoReverseProxy(mountPrefix string, upstreamPrefix string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		target, err := mihomoAPIURL(api.mihomoAPIAddr)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Director = func(out *http.Request) {
			out.URL.Scheme = target.Scheme
			out.URL.Host = target.Host
			out.URL.Path = joinMihomoProxyPath(upstreamPrefix, strings.TrimPrefix(ctx.Request.URL.Path, mountPrefix))
			out.Host = target.Host
			out.Header.Set("X-Forwarded-Host", ctx.Request.Host)
			out.Header.Set("X-Forwarded-Proto", forwardedProto(ctx.Request))
		}
		proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
			writeHTTPError(w, err, http.StatusBadGateway)
		}
		proxy.ModifyResponse = redactMihomoControllerErrorResponse
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
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
