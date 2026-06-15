package app

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

func (c leaseCoordinator) preparedAcquireRunner(advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.PreparedAcquireRunner {
	accountLockedFactory := leaseAccountLockedAcquireRunnerFactory{
		deps:           c.deps,
		advertisedHost: advertisedHost,
		request:        req,
		retire:         c.leaseRouteRetirer(),
		reuse:          c.concurrencySlotRefreshRunner(),
	}
	action := leaseapp.SettingsPreparedAcquireAction[*runtimeSettingsFile]{
		Load:    c.deps.settings.load,
		Request: req,
		EgressProfiles: func(settings *runtimeSettingsFile) []*proxyruntimev1.EgressProfileSettings {
			return settings.GetEgressProfiles()
		},
		NewRunner:      accountLockedFactory.New,
		MapPolicyError: leaseProfilePolicyError,
	}
	return leaseapp.PreparedAcquireRunner{
		Locks:  c.deps.locks,
		Action: action.Run,
	}
}
