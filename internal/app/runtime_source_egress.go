package app

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func sourcePlaneEgressProfileExit(exit *proxyruntimev1.EgressProfileExitSettings) sourceplane.EgressProfileExit {
	return sourceplane.EgressProfileExit{
		Kind:           egressProfileExitKind(exit.GetKind()),
		ProviderID:     strings.TrimSpace(exit.GetDynamicProviderId()),
		ResourceID:     strings.TrimSpace(exit.GetMihomoNode().GetResourceId()),
		NodeID:         strings.TrimSpace(exit.GetMihomoNode().GetNodeId()),
		HealthCheckURL: strings.TrimSpace(exit.GetHealthCheckUrl()),
		HealthInterval: protoDuration(exit.GetHealthInterval(), 300*time.Second),
		HealthTimeout:  protoDuration(exit.GetHealthTimeout(), 5*time.Second),
		ExpectedStatus: defaultExpectedStatus(exit.GetExpectedStatus()),
	}
}

func egressProfileLineKind(kind proxyruntimev1.EgressProfileLineKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE:
		return "mihomo_node"
	default:
		return "direct"
	}
}

func egressProfileExitKind(kind proxyruntimev1.EgressProfileExitKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DIRECT:
		return "direct"
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
		return "static_ip"
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
		return "dynamic_ip"
	default:
		return ""
	}
}

func sourcePlaneEgressProfileLine(line *proxyruntimev1.EgressProfileLineSettings) sourceplane.EgressProfileLine {
	return sourceplane.EgressProfileLine{
		Kind:           egressProfileLineKind(line.GetKind()),
		ResourceID:     strings.TrimSpace(line.GetMihomoNode().GetResourceId()),
		NodeID:         strings.TrimSpace(line.GetMihomoNode().GetNodeId()),
		HealthCheckURL: strings.TrimSpace(line.GetHealthCheckUrl()),
		HealthInterval: protoDuration(line.GetHealthInterval(), 300*time.Second),
		HealthTimeout:  protoDuration(line.GetHealthTimeout(), 5*time.Second),
		ExpectedStatus: defaultExpectedStatus(line.GetExpectedStatus()),
	}
}

func sourcePlaneEgressProfiles(settings *runtimeSettingsFile) []sourceplane.EgressProfile {
	settings = kernel.NormalizeRuntimeSettings(settings)
	out := make([]sourceplane.EgressProfile, 0, len(settings.GetEgressProfiles()))
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() {
			continue
		}
		out = append(out, sourceplane.EgressProfile{
			ID:          profile.GetProfileId(),
			DisplayName: profile.GetDisplayName(),
			Enabled:     profile.GetEnabled(),
			Line:        sourcePlaneEgressProfileLine(profile.GetLine()),
			Exit:        sourcePlaneEgressProfileExit(profile.GetExit()),
		})
	}
	return out
}
