package app

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
