package app

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

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
