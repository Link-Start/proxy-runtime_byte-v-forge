package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAccountLockedAcquireRunnerFactoryRequired = errors.New("account locked acquire runner factory is required")

type SettingsEgressProfilesFunc[T any] func(T) []*proxyruntimev1.EgressProfileSettings
type SettingsIngressRulesFunc[T any] func(T) []*proxyruntimev1.ProxyIngressRuleSettings

type AccountLockedAcquireRunnerFactory[T any] func(context.Context, T) AccountLockedAcquireRunner

type AcquirePolicyErrorMapper func(error) error

type SettingsPreparedAcquireAction[T any] struct {
	Load           SettingsLoader[T]
	Request        *proxyruntimev1.AcquireProxyLeaseRequest
	EgressProfiles SettingsEgressProfilesFunc[T]
	IngressRules   SettingsIngressRulesFunc[T]
	NewRunner      AccountLockedAcquireRunnerFactory[T]
	MapPolicyError AcquirePolicyErrorMapper
}

func (a SettingsPreparedAcquireAction[T]) Run(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
	if a.Load == nil {
		return nil, ErrSettingsLoaderRequired
	}
	settings, err := a.Load(ctx)
	if err != nil {
		return nil, err
	}
	if a.NewRunner == nil {
		return nil, ErrAccountLockedAcquireRunnerFactoryRequired
	}
	egressProfiles := a.egressProfiles(settings)
	ingressRules := a.ingressRules(settings)
	ResolveAcquireRequestAccountID(egressProfiles, ingressRules, a.Request)
	runner := a.NewRunner(ctx, settings)
	lease, err := runner.Run(ctx, AccountLockedAcquireRunnerInput{
		Request:        a.Request,
		EgressProfiles: egressProfiles,
	})
	if err != nil && IsAcquirePolicyError(err) {
		return nil, a.mapPolicyError(err)
	}
	return lease, err
}

func (a SettingsPreparedAcquireAction[T]) egressProfiles(settings T) []*proxyruntimev1.EgressProfileSettings {
	if a.EgressProfiles == nil {
		return nil
	}
	return a.EgressProfiles(settings)
}

func (a SettingsPreparedAcquireAction[T]) ingressRules(settings T) []*proxyruntimev1.ProxyIngressRuleSettings {
	if a.IngressRules == nil {
		return nil
	}
	return a.IngressRules(settings)
}

func (a SettingsPreparedAcquireAction[T]) mapPolicyError(err error) error {
	if err == nil || a.MapPolicyError == nil {
		return err
	}
	return a.MapPolicyError(err)
}
