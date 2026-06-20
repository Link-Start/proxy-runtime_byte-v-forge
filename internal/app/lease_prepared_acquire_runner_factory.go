package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type leasePreparedAcquireRunnerFactory struct {
	deps           leaseCoordinatorDependencies
	advertisedHost string
	request        *proxyruntimev1.AcquireProxyLeaseRequest
	retire         leaseapp.LeaseRouteRetirer
	reuse          leaseapp.RefreshConcurrencySlotRunner
}

func (f leasePreparedAcquireRunnerFactory) New() leaseapp.PreparedAcquireRunner {
	accountLockedFactory := leaseAccountLockedAcquireRunnerFactory{
		deps:           f.deps,
		advertisedHost: f.advertisedHost,
		request:        f.request,
		retire:         f.retire,
		reuse:          f.reuse,
	}
	action := leaseapp.SettingsPreparedAcquireAction[*runtimeSettingsFile]{
		Load:           f.deps.settings.Load,
		Request:        f.request,
		EgressProfiles: leaseSettingsEgressProfiles,
		IngressRules:   leaseSettingsIngressRules,
		NewRunner:      accountLockedFactory.New,
		MapPolicyError: leaseProfilePolicyError,
	}
	return leaseapp.PreparedAcquireRunner{
		Locks:  f.deps.locks,
		Action: action.Run,
	}
}

func (c leaseCoordinator) preparedAcquireRunner(advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.PreparedAcquireRunner {
	return leasePreparedAcquireRunnerFactory{
		deps:           c.deps,
		advertisedHost: advertisedHost,
		request:        req,
		retire:         c.leaseRouteRetirer(),
		reuse:          c.concurrencySlotRefreshRunner(),
	}.New()
}

func leaseSettingsEgressProfiles(settings *runtimeSettingsFile) []*proxyruntimev1.EgressProfileSettings {
	return settings.GetEgressProfiles()
}

func leaseSettingsIngressRules(settings *runtimeSettingsFile) []*proxyruntimev1.ProxyIngressRuleSettings {
	return settings.GetIngressRules()
}
