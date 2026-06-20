package app

import (
	"context"
	"io"
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	settingsdomain "github.com/byte-v-forge/proxy-runtime/internal/app/settings/domain"
)

func (r *Runtime) runEdgeCanary(ctx context.Context, client *http.Client, settings *runtimeSettingsFile) proxycheck.EdgeCanaryOutcome {
	edgeCanary := settings.GetEdgeCanary()
	target := strings.TrimSpace(edgeCanary.GetUrl())
	token := ""
	if r.store != nil && appcore.SecretRefConfigured(edgeCanary.GetTokenSecretRef()) {
		resolved, err := r.store.ResolveSecret(ctx, edgeCanary.GetTokenSecretRef())
		if err == nil {
			token = strings.TrimSpace(resolved)
		}
	}
	if !settingsdomain.EdgeCanaryEnabled(edgeCanary) || target == "" || token == "" {
		return proxycheck.EdgeCanaryOutcome{
			Level:        proxyruntimev1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNSUPPORTED,
			Signal:       proxyruntimev1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_ACCESS_UNSUPPORTED,
			ErrorMessage: "edge access canary is not configured",
		}
	}
	canaryCtx, cancel := context.WithTimeout(ctx, r.cfg.EdgeCanaryTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(canaryCtx, http.MethodGet, target, nil)
	if err != nil {
		return edgeUnavailableOutcome()
	}
	req.Close = true
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "close")
	req.Header.Set("X-Canary-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		return edgeUnavailableOutcome()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return edgeUnavailableOutcome()
	}
	if edgeChallengeDetected(resp.Header, body) {
		return proxycheck.EdgeCanaryOutcome{
			Level:  proxyruntimev1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_CHALLENGE_LIKELY,
			Score:  85,
			Signal: proxyruntimev1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_CHALLENGE_DETECTED,
		}
	}
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return proxycheck.EdgeCanaryOutcome{
			Level:  proxyruntimev1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH,
			Score:  70,
			Signal: proxyruntimev1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_RATE_LIMIT_DETECTED,
		}
	case resp.StatusCode == http.StatusForbidden:
		return proxycheck.EdgeCanaryOutcome{
			Level:  proxyruntimev1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_BLOCK_LIKELY,
			Score:  90,
			Signal: proxyruntimev1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_BLOCK_DETECTED,
		}
	case resp.StatusCode >= 500 || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized:
		return edgeUnavailableOutcome()
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return proxycheck.EdgeCanaryOutcome{
			Level:  proxyruntimev1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_LOW,
			Score:  0,
			Signal: proxyruntimev1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_ACCESS_PASSED,
		}
	default:
		return edgeUnavailableOutcome()
	}
}

func edgeUnavailableOutcome() proxycheck.EdgeCanaryOutcome {
	return proxycheck.EdgeCanaryOutcome{
		Level:        proxyruntimev1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN,
		Signal:       proxyruntimev1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_UNAVAILABLE,
		ErrorMessage: "edge access check unavailable",
	}
}

func edgeChallengeDetected(header http.Header, body []byte) bool {
	if strings.EqualFold(strings.TrimSpace(header.Get("cf-mitigated")), "challenge") {
		return true
	}
	text := strings.ToLower(string(body))
	for _, hint := range []string{"challenge-platform", "cf-chl", "just a moment"} {
		if strings.Contains(text, hint) {
			return true
		}
	}
	return false
}
