package app

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

type leaseSelectedAcquireAttemptRunnerFactory struct {
	deps           leaseCoordinatorDependencies
	settings       *runtimeSettingsFile
	advertisedHost string
	request        *proxyruntimev1.AcquireProxyLeaseRequest
}

func (f leaseSelectedAcquireAttemptRunnerFactory) New(selection leaseapp.DynamicIPSelection) leaseapp.SelectedAcquireAttemptRunner {
	providerRunnerFactory := leaseProviderAccountAcquireRunnerFactory{
		deps:           f.deps,
		settings:       f.settings,
		advertisedHost: f.advertisedHost,
		request:        f.request,
		selectionPlan:  selection.Plan,
	}
	action := leaseapp.SelectedAttemptProviderAccountAction{
		Selection: selection,
		Request:   f.request,
		NewRunner: providerRunnerFactory.New,
		MapError:  providerSessionAcquireError,
	}
	return leaseapp.SelectedAcquireAttemptRunner{
		Store:          f.deps.store,
		IDs:            f.deps.ids,
		Limiter:        f.deps.providerConcurrency,
		Locks:          f.deps.locks,
		DefaultTTL:     leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:      providerAccountConcurrencyTTLBuffer,
		ReleaseTimeout: leaseAcquireSlotReleaseTimeout,
		Limit:          f.limit,
		Action:         action.Run,
	}
}

func (f leaseSelectedAcquireAttemptRunnerFactory) limit(selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	return dynamicProviderConcurrencyLimit(f.settings, leaseapp.SelectedDynamicProviderID(selectionPlan), policy)
}
