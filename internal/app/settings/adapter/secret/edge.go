package secret

import (
	"context"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func edgeCanaryFromRequest(ctx context.Context, writer secretref.Writer, req *proxygatewayv1.ProxyEdgeCanarySettings, current *proxygatewayv1.ProxyEdgeCanarySettings) (*proxygatewayv1.ProxyEdgeCanarySettings, error) {
	if req == nil {
		return cloneEdgeCanary(current), nil
	}
	settings := &proxygatewayv1.ProxyEdgeCanarySettings{
		Enabled: req.GetEnabled(),
		Url:     appcore.FirstNonEmpty(req.GetUrl(), current.GetUrl()),
	}
	if ref := appcore.CloneSecretRef(req.GetTokenSecretRef(), "proxy-gateway", "edge_canary_token"); ref != nil {
		settings.TokenSecretRef = ref
		return settings, nil
	}
	if rawToken := strings.TrimSpace(req.GetTokenValue()); rawToken != "" {
		ref, err := WriteRuntimeSecret(ctx, writer, rawToken, secretref.StableID("proxy-gateway-edge-canary-token", "default"), "edge_canary_token")
		if err != nil {
			return nil, err
		}
		settings.TokenSecretRef = ref
		return settings, nil
	}
	if req.GetClearToken() {
		settings.TokenSecretRef = nil
		return settings, nil
	}
	settings.TokenSecretRef = appcore.CloneSecretRef(current.GetTokenSecretRef(), "proxy-gateway", "edge_canary_token")
	return settings, nil
}

func cloneEdgeCanary(in *proxygatewayv1.ProxyEdgeCanarySettings) *proxygatewayv1.ProxyEdgeCanarySettings {
	if in == nil {
		return nil
	}
	return &proxygatewayv1.ProxyEdgeCanarySettings{
		Url:            in.GetUrl(),
		TokenSecretRef: appcore.CloneSecretRef(in.GetTokenSecretRef(), "proxy-gateway", "edge_canary_token"),
		ClearToken:     in.GetClearToken(),
		Enabled:        in.GetEnabled(),
		TokenValue:     "",
	}
}
