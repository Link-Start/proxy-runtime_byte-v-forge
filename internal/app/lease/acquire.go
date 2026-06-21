package lease

import (
	"context"
	"errors"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
)

type PreparedAcquireRunner struct {
	Locks  LockManager
	Action AccountLeaseAction
}

type PreparedAcquireRunnerInput struct {
	Request *proxygatewayv1.AcquireProxyLeaseRequest
}

func (r PreparedAcquireRunner) Run(ctx context.Context, input PreparedAcquireRunnerInput) (*proxygatewayv1.ProxyDynamicLease, error) {
	if err := PrepareAcquireRequest(input.Request); err != nil {
		return nil, err
	}
	return RunAccountLeaseAction(ctx, r.Locks, input.Request.GetAccountId(), r.Action)
}

func IsAcquireRequestError(err error) bool {
	return errors.Is(err, ErrAcquireRequestRequired) ||
		errors.Is(err, ErrAcquireAccountIDRequired)
}

type AccountLockedAcquireRunner struct {
	Store               OrchestrationStore
	Clock               clock.Clock
	PlaygroundAccountID string
	PlaygroundUsername  string
	Reuse               ExistingActiveLeaseAction
	Replace             ExistingActiveLeaseAction
	RunAttempt          AcquireAttemptRunner
	Retry               AcquireAttemptRetryPolicy
	Observe             AcquireAttemptFailureObserver
}

type AccountLockedAcquireRunnerInput struct {
	Request        *proxygatewayv1.AcquireProxyLeaseRequest
	EgressProfiles []*proxygatewayv1.EgressProfileSettings
}

func (r AccountLockedAcquireRunner) Run(ctx context.Context, input AccountLockedAcquireRunnerInput) (*proxygatewayv1.ProxyDynamicLease, error) {
	selectionPolicy, err := ApplyAcquireRequestPolicies(input.Request, input.EgressProfiles)
	if err != nil {
		return nil, err
	}
	existing, handled, err := HandleExistingActiveLease(ctx, ExistingActiveLeaseInput{
		Store:               r.Store,
		Request:             input.Request,
		Now:                 r.now().UTC(),
		PlaygroundAccountID: r.PlaygroundAccountID,
		PlaygroundUsername:  r.PlaygroundUsername,
		Reuse:               r.Reuse,
		Replace:             r.Replace,
	})
	if err != nil {
		return nil, err
	}
	if handled {
		return existing, nil
	}
	return RunAcquireAttempts(ctx, input.Request, selectionPolicy, r.RunAttempt, r.Retry, r.Observe)
}

func (r AccountLockedAcquireRunner) now() time.Time {
	if r.Clock != nil {
		return r.Clock.Now()
	}
	return time.Now()
}

func IsAcquirePolicyError(err error) bool {
	return errors.Is(err, ErrProfileDynamicIPNotConfigured) ||
		errors.Is(err, ErrProfileLeaseRequiresSticky) ||
		errors.Is(err, ErrRequestRequiresStickyDynamicIP)
}

var (
	ErrDynamicIPSelectorRequired              = errors.New("dynamic IP selector is required")
	ErrSelectedAcquireAttemptRunnerFactoryNil = errors.New("selected acquire attempt runner factory is required")
)

type DynamicIPSelectorFunc func(context.Context, *proxygatewayv1.AcquireProxyLeaseRequest) (DynamicIPSelection, error)

type SelectedAcquireAttemptRunnerFactory func(DynamicIPSelection) SelectedAcquireAttemptRunner

type AcquireAttemptErrorMapper func(error) error

type DynamicAcquireAttemptRunner struct {
	Select            DynamicIPSelectorFunc
	NewSelectedRunner SelectedAcquireAttemptRunnerFactory
	MapSelectionError AcquireAttemptErrorMapper
	MapAttemptError   AcquireAttemptErrorMapper
}

func (r DynamicAcquireAttemptRunner) Run(ctx context.Context, req *proxygatewayv1.AcquireProxyLeaseRequest, policy *proxygatewayv1.ProxySessionPolicy) (*proxygatewayv1.ProxyDynamicLease, error) {
	if r.Select == nil {
		return nil, ErrDynamicIPSelectorRequired
	}
	selection, err := r.Select(ctx, req)
	if err != nil {
		return nil, mapAcquireAttemptError(r.MapSelectionError, err)
	}
	if r.NewSelectedRunner == nil {
		return nil, ErrSelectedAcquireAttemptRunnerFactoryNil
	}
	runner := r.NewSelectedRunner(selection)
	lease, err := runner.Run(ctx, SelectedAcquireAttemptRunnerInput{
		SelectionPlan: selection.Plan,
		Policy:        policy,
	})
	if err != nil {
		return nil, mapAcquireAttemptError(r.MapAttemptError, err)
	}
	return lease, nil
}

func mapAcquireAttemptError(mapper AcquireAttemptErrorMapper, err error) error {
	if err == nil || mapper == nil {
		return err
	}
	return mapper(err)
}

var ErrAccountLockedAcquireRunnerFactoryRequired = errors.New("account locked acquire runner factory is required")

type SettingsEgressProfilesFunc[T any] func(T) []*proxygatewayv1.EgressProfileSettings
type SettingsIngressRulesFunc[T any] func(T) []*proxygatewayv1.ProxyIngressRuleSettings

type AccountLockedAcquireRunnerFactory[T any] func(context.Context, T) AccountLockedAcquireRunner

type AcquirePolicyErrorMapper func(error) error

type SettingsPreparedAcquireAction[T any] struct {
	Load           SettingsLoader[T]
	Request        *proxygatewayv1.AcquireProxyLeaseRequest
	EgressProfiles SettingsEgressProfilesFunc[T]
	IngressRules   SettingsIngressRulesFunc[T]
	NewRunner      AccountLockedAcquireRunnerFactory[T]
	MapPolicyError AcquirePolicyErrorMapper
}

func (a SettingsPreparedAcquireAction[T]) Run(ctx context.Context) (*proxygatewayv1.ProxyDynamicLease, error) {
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

func (a SettingsPreparedAcquireAction[T]) egressProfiles(settings T) []*proxygatewayv1.EgressProfileSettings {
	if a.EgressProfiles == nil {
		return nil
	}
	return a.EgressProfiles(settings)
}

func (a SettingsPreparedAcquireAction[T]) ingressRules(settings T) []*proxygatewayv1.ProxyIngressRuleSettings {
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
