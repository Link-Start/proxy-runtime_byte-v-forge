package app

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func sourcePlaneEgressProfiles(settings *runtimeSettingsFile) []sourceplane.EgressProfile {
	settings = normalizeRuntimeSettings(settings)
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
