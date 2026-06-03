package app

import proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func edgeCanaryFromRequest(req *proxyruntimev1.ProxyEdgeCanarySettings, current *proxyruntimev1.ProxyEdgeCanarySettings) *proxyruntimev1.ProxyEdgeCanarySettings {
	if req == nil {
		return cloneEdgeCanary(current)
	}
	settings := &proxyruntimev1.ProxyEdgeCanarySettings{
		Enabled:        req.GetEnabled(),
		Url:            firstNonEmpty(req.GetUrl(), current.GetUrl()),
		TokenSecretRef: cloneSecretRef(req.GetTokenSecretRef(), "proxy-runtime", "edge_canary_token"),
	}
	switch {
	case secretRefValue(settings.GetTokenSecretRef()) != "":
	case req.GetClearToken():
		settings.TokenSecretRef = nil
	default:
		settings.TokenSecretRef = cloneSecretRef(current.GetTokenSecretRef(), "proxy-runtime", "edge_canary_token")
	}
	return settings
}

func edgeCanaryEnabled(settings *proxyruntimev1.ProxyEdgeCanarySettings) bool {
	return settings != nil && settings.GetEnabled()
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
	}
}
