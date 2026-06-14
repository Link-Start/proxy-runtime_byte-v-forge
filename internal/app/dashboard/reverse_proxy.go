package dashboard

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"

	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
)

type ProxyErrorWriter func(http.ResponseWriter, error, int)

type ReverseProxyOptions struct {
	APIAddr        string
	MountPrefix    string
	UpstreamPrefix string
	AuthToken      string
	WriteError     ProxyErrorWriter
}

type proxyContextKey string

const proxyRequestPathKey proxyContextKey = "request_path"

func NewReverseProxy(opts ReverseProxyOptions) (http.Handler, error) {
	target, err := APIURL(opts.APIAddr)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(out *http.Request) {
		requestPath := out.URL.Path
		requestHost := out.Host
		*out = *out.WithContext(context.WithValue(out.Context(), proxyRequestPathKey, requestPath))
		out.URL.Scheme = target.Scheme
		out.URL.Host = target.Host
		out.URL.Path = JoinProxyPath(opts.UpstreamPrefix, strings.TrimPrefix(requestPath, opts.MountPrefix))
		out.Host = target.Host
		out.Header.Set("X-Forwarded-Host", requestHost)
		out.Header.Set("X-Forwarded-Proto", httpapi.ForwardedProto(out))
		InjectControllerAuthorization(out, requestPath, opts.AuthToken)
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		writeProxyError(w, err, opts.WriteError)
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		ApplyCacheHeaders(resp, proxyRequestPath(resp.Request))
		return RedactControllerErrorResponse(resp)
	}
	return proxy, nil
}

func InjectControllerAuthorization(out *http.Request, requestPath string, token string) {
	if !httpapi.PathInPrefix(requestPath, "/mihomo/controller") {
		return
	}
	ApplyControllerAuthorization(out, token)
}

func ApplyControllerAuthorization(out *http.Request, token string) {
	if out == nil {
		return
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	out.Header.Set("Authorization", "Bearer "+token)
	if out.URL != nil {
		out.URL.RawQuery = ControllerUpstreamRawQuery(out.URL.RawQuery)
	}
}

func proxyRequestPath(req *http.Request) string {
	if req != nil {
		if requestPath, ok := req.Context().Value(proxyRequestPathKey).(string); ok {
			return requestPath
		}
		if req.URL != nil {
			return req.URL.Path
		}
	}
	return ""
}

func writeProxyError(w http.ResponseWriter, err error, writeError ProxyErrorWriter) {
	if writeError != nil {
		writeError(w, err, http.StatusBadGateway)
		return
	}
	http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
}
