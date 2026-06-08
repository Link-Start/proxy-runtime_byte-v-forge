package app

import (
	"context"
	"errors"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func edgeCanaryFromRequest(ctx context.Context, writer secretref.Writer, req *proxyruntimev1.ProxyEdgeCanarySettings, current *proxyruntimev1.ProxyEdgeCanarySettings) (*proxyruntimev1.ProxyEdgeCanarySettings, error) {
	if req == nil {
		return cloneEdgeCanary(current), nil
	}
	settings := &proxyruntimev1.ProxyEdgeCanarySettings{
		Enabled: req.GetEnabled(),
		Url:     firstNonEmpty(req.GetUrl(), current.GetUrl()),
	}
	if ref := cloneSecretRef(req.GetTokenSecretRef(), "proxy-runtime", "edge_canary_token"); ref != nil {
		settings.TokenSecretRef = ref
		return settings, nil
	}
	if rawToken := strings.TrimSpace(req.GetTokenValue()); rawToken != "" {
		ref, err := writeRuntimeSecret(ctx, writer, rawToken, secretref.StableID("proxy-runtime-edge-canary-token", "default"), "edge_canary_token")
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
	settings.TokenSecretRef = cloneSecretRef(current.GetTokenSecretRef(), "proxy-runtime", "edge_canary_token")
	return settings, nil
}

func edgeCanaryEnabled(settings *proxyruntimev1.ProxyEdgeCanarySettings) bool {
	return settings != nil && settings.GetEnabled()
}

func writeRuntimeSecret(ctx context.Context, writer secretref.Writer, raw string, secretID string, purpose string) (*commonv1.SecretRef, error) {
	if writer == nil {
		return nil, errors.New("proxy-runtime secret store is required")
	}
	return writer.WriteSecret(ctx, secretref.WriteRequest{
		SecretID: secretID,
		Provider: "proxy-runtime",
		Purpose:  purpose,
		Value:    raw,
	})
}

func cloneEdgeCanary(in *proxyruntimev1.ProxyEdgeCanarySettings) *proxyruntimev1.ProxyEdgeCanarySettings {
	if in == nil {
		return nil
	}
	return &proxyruntimev1.ProxyEdgeCanarySettings{
		Url:            in.GetUrl(),
		TokenSecretRef: cloneSecretRef(in.GetTokenSecretRef(), "proxy-runtime", "edge_canary_token"),
		ClearToken:     in.GetClearToken(),
		Enabled:        in.GetEnabled(),
		TokenValue:     "",
	}
}
