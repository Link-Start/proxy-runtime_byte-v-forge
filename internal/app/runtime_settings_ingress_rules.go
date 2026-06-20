package app

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func sourcePlaneProxyUserRoutes(settings *runtimeSettingsFile) []dataplane.ProxyUserRoute {
	settings = kernel.NormalizeRuntimeSettings(settings)
	out := make([]dataplane.ProxyUserRoute, 0, len(settings.GetIngressRules()))
	for _, rule := range settings.GetIngressRules() {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetUsername()) == "" {
			continue
		}
		out = append(out, dataplane.ProxyUserRoute{
			ID:        rule.GetRuleId(),
			Username:  rule.GetUsername(),
			Password:  rule.GetPasswordValue(),
			Route:     config.ListenerRouteProfile,
			ProfileID: rule.GetProfileId(),
		})
	}
	return out
}
