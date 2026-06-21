package app

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
)

type leasePreparedAcquireRunnerFactory struct {
	deps           leaseCoordinatorDependencies
	advertisedHost string
	request        *proxygatewayv1.AcquireProxyLeaseRequest
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

func (c leaseCoordinator) preparedAcquireRunner(advertisedHost string, req *proxygatewayv1.AcquireProxyLeaseRequest) leaseapp.PreparedAcquireRunner {
	return leasePreparedAcquireRunnerFactory{
		deps:           c.deps,
		advertisedHost: advertisedHost,
		request:        req,
		retire:         c.leaseRouteRetirer(),
		reuse:          c.concurrencySlotRefreshRunner(),
	}.New()
}

func leaseSettingsEgressProfiles(settings *runtimeSettingsFile) []*proxygatewayv1.EgressProfileSettings {
	return settings.GetEgressProfiles()
}

func leaseSettingsIngressRules(settings *runtimeSettingsFile) []*proxygatewayv1.ProxyIngressRuleSettings {
	return settings.GetIngressRules()
}
