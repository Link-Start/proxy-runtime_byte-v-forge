package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const dynamicIPEndpointHealthWindow = 6 * time.Hour

const dynamicIPEndpointHealthLimit = 200

type dynamicIPEndpointHealth struct {
	success int
	failure int
}

func (p *dynamicIPSelector) dynamicIPEndpointHealthScores(ctx context.Context) map[string]int {
	if p == nil || p.store == nil {
		return nil
	}
	leases, err := p.store.RecentLeaseFacts(ctx, time.Now().UTC().Add(-dynamicIPEndpointHealthWindow), dynamicIPEndpointHealthLimit)
	if err != nil {
		p.logger.Warn("load dynamic IP endpoint health facts failed", "error", err)
		return nil
	}
	return dynamicIPEndpointHealthScoresFromLeases(leases)
}

func applyDynamicIPEndpointHealthScores(candidates []scoredDynamicIPEndpointCandidate, scores map[string]int) {
	if len(scores) == 0 {
		return
	}
	for index := range candidates {
		endpointID := candidates[index].proto.GetEndpointId()
		candidates[index].score += scores[endpointID]
	}
}

func dynamicIPEndpointHealthScoresFromLeases(leases []*proxyruntimev1.ProxyDynamicLease) map[string]int {
	stats := map[string]dynamicIPEndpointHealth{}
	for _, lease := range leases {
		endpointID := leaseapp.EndpointID(lease)
		if endpointID == "" {
			continue
		}
		stat := stats[endpointID]
		switch {
		case leaseapp.HasFailedStatus(lease):
			stat.failure++
		case leaseapp.HasSuccessfulAttemptStatus(lease):
			stat.success++
		}
		stats[endpointID] = stat
	}
	out := map[string]int{}
	for endpointID, stat := range stats {
		out[endpointID] = min(stat.success*20, 120) - min(stat.failure*150, 450)
	}
	return out
}
