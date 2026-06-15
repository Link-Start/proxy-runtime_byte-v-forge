package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAccountLockedAcquireRunnerFactoryRequired = errors.New("account locked acquire runner factory is required")

type SettingsEgressProfilesFunc[T any] func(T) []*proxyruntimev1.EgressProfileSettings

type AccountLockedAcquireRunnerFactory[T any] func(context.Context, T) AccountLockedAcquireRunner

type AcquirePolicyErrorMapper func(error) error

type SettingsPreparedAcquireAction[T any] struct {
	Load           SettingsLoader[T]
	Request        *proxyruntimev1.AcquireProxyLeaseRequest
	EgressProfiles SettingsEgressProfilesFunc[T]
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
	runner := a.NewRunner(ctx, settings)
	lease, err := runner.Run(ctx, AccountLockedAcquireRunnerInput{
		Request:        a.Request,
		EgressProfiles: a.egressProfiles(settings),
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

func (a SettingsPreparedAcquireAction[T]) mapPolicyError(err error) error {
	if err == nil || a.MapPolicyError == nil {
		return err
	}
	return a.MapPolicyError(err)
}
