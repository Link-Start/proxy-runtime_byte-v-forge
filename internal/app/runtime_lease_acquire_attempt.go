package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

func (c leaseCoordinator) acquireLeaseAttempt(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile) (*proxyruntimev1.ProxyDynamicLease, error) {
	selection, err := c.deps.dynamicIPSelector.selectDynamicIPEndpoint(ctx, req)
	if err != nil {
		return nil, failedPrecondition("no dynamic IP endpoint candidate", err)
	}
	runner := c.selectedAcquireAttemptRunner(settings, advertisedHost, req, selection)
	lease, err := runner.Run(ctx, leaseapp.SelectedAcquireAttemptRunnerInput{
		SelectionPlan: selection.plan,
		Policy:        req.GetPolicy(),
	})
	if err != nil {
		return nil, acquireAttemptSlotError(err)
	}
	return lease, nil
}
