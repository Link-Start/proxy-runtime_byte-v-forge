package lease

import (
	"context"
	"errors"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

const (
	LabelAccountID                        = "account_id"
	LabelLeaseID                          = "lease_id"
	LabelPurpose                          = "purpose"
	LabelSessionID                        = "session_id"
	LabelProviderAccountID                = "provider_account_id"
	LabelProviderAccountConcurrencyHolder = "provider_account_concurrency_holder"
	LabelDynamicProviderID                = "dynamic_provider_id"
	LabelDynamicIPEndpointID              = "dynamic_ip_endpoint_id"
	LabelSelectionID                      = "selection_id"
	LabelAttempt                          = "attempt"
)

func ConcurrencyMode(policy *proxygatewayv1.ProxySessionPolicy) proxygatewayv1.ProxySessionMode {
	if policy == nil {
		return proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
	}
	if policy.GetMode() == proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING || policy.GetRotationMode() == proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST {
		return proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING
	}
	return proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
}

func ConcurrencyModeText(policy *proxygatewayv1.ProxySessionPolicy) string {
	if ConcurrencyMode(policy) == proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return "rotating"
	}
	return "sticky"
}

func ConcurrencySlotTTL(policy *proxygatewayv1.ProxySessionPolicy, defaultTTL time.Duration, buffer time.Duration) time.Duration {
	ttl := defaultTTL
	if policy != nil && policy.GetStickyTtl() != nil && policy.GetStickyTtl().AsDuration() > 0 {
		ttl = policy.GetStickyTtl().AsDuration()
	}
	return ttl + buffer
}

func ConcurrencyPolicy(lease *proxygatewayv1.ProxyDynamicLease) *proxygatewayv1.ProxySessionPolicy {
	if lease == nil {
		return nil
	}
	if policy := lease.GetSession().GetPolicy(); policy != nil {
		return policy
	}
	return &proxygatewayv1.ProxySessionPolicy{RotationMode: lease.GetEgress().GetRotationMode()}
}

func ConcurrencyHolder(lease *proxygatewayv1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if holder := strings.TrimSpace(lease.GetEgress().GetLabels()[LabelProviderAccountConcurrencyHolder]); holder != "" {
		return holder
	}
	if holder := strings.TrimSpace(lease.GetSession().GetLabels()[LabelProviderAccountConcurrencyHolder]); holder != "" {
		return holder
	}
	return HolderForLeaseID(lease.GetLeaseId())
}

func HolderForLeaseID(leaseID string) string {
	if leaseID = strings.TrimSpace(leaseID); leaseID != "" {
		return "lease:" + leaseID
	}
	return ""
}

func DynamicProviderID(lease *proxygatewayv1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetEgress().GetLabels()[LabelDynamicProviderID]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetSession().GetLabels()[LabelDynamicProviderID]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if endpoint := lease.GetSelectionPlan().GetSelectedEndpoint(); endpoint != nil {
		return endpoint.GetDynamicProviderId()
	}
	return ""
}

type RefreshConcurrencySlotInput struct {
	Store      OrchestrationStore
	Limiter    ProviderAccountConcurrencyLimiter
	Lease      *proxygatewayv1.ProxyDynamicLease
	Limit      uint32
	DefaultTTL time.Duration
	TTLBuffer  time.Duration
}

type RefreshConcurrencySlotLimitFunc func(context.Context, *proxygatewayv1.ProxyDynamicLease, *proxygatewayv1.ProxySessionPolicy) (uint32, error)

type RefreshConcurrencySlotRunner struct {
	Store      OrchestrationStore
	Limiter    ProviderAccountConcurrencyLimiter
	DefaultTTL time.Duration
	TTLBuffer  time.Duration
	Limit      RefreshConcurrencySlotLimitFunc
}

func (r RefreshConcurrencySlotRunner) Refresh(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	if !NeedsConcurrencySlotRefresh(lease, r.Limiter) {
		return nil
	}
	policy := ConcurrencyPolicy(lease)
	limit, err := r.limit(ctx, lease, policy)
	if err != nil {
		return err
	}
	return RefreshConcurrencySlot(ctx, RefreshConcurrencySlotInput{
		Store:      r.Store,
		Limiter:    r.Limiter,
		Lease:      lease,
		Limit:      limit,
		DefaultTTL: r.DefaultTTL,
		TTLBuffer:  r.TTLBuffer,
	})
}

func (r RefreshConcurrencySlotRunner) limit(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease, policy *proxygatewayv1.ProxySessionPolicy) (uint32, error) {
	if r.Limit == nil {
		return 0, nil
	}
	return r.Limit(ctx, lease, policy)
}

func RefreshConcurrencySlot(ctx context.Context, input RefreshConcurrencySlotInput) error {
	if !NeedsConcurrencySlotRefresh(input.Lease, input.Limiter) {
		return nil
	}
	providerAccountID := strings.TrimSpace(input.Lease.GetProviderAccountId())
	holder := strings.TrimSpace(ConcurrencyHolder(input.Lease))
	account, err := input.Store.ProviderAccount(ctx, providerAccountID)
	if err != nil {
		return err
	}
	policy := ConcurrencyPolicy(input.Lease)
	_, err = AcquireProviderAccountConcurrencySlot(
		ctx,
		input.Limiter,
		account.GetAccountId(),
		input.Limit,
		policy,
		holder,
		ConcurrencySlotTTL(policy, input.DefaultTTL, input.TTLBuffer),
	)
	return err
}

func NeedsConcurrencySlotRefresh(lease *proxygatewayv1.ProxyDynamicLease, limiter ProviderAccountConcurrencyLimiter) bool {
	return lease != nil &&
		limiter != nil &&
		strings.TrimSpace(lease.GetProviderAccountId()) != "" &&
		strings.TrimSpace(ConcurrencyHolder(lease)) != ""
}

var ErrTemporaryConcurrencyActionRequired = errors.New("temporary concurrency action is required")

type TemporaryConcurrencyAction func(context.Context) error

type TemporaryConcurrencySlotInput struct {
	Slot           ProviderAccountConcurrencySlot
	ReleaseTimeout time.Duration
	Action         TemporaryConcurrencyAction
}

func RunTemporaryConcurrencySlot(ctx context.Context, input TemporaryConcurrencySlotInput) error {
	if input.Action == nil {
		return ErrTemporaryConcurrencyActionRequired
	}
	keep := false
	defer func() {
		releaseCtx := context.WithoutCancel(ctx)
		if input.ReleaseTimeout > 0 {
			var cancel context.CancelFunc
			releaseCtx, cancel = context.WithTimeout(releaseCtx, input.ReleaseTimeout)
			defer cancel()
		}
		_ = ReleaseConcurrencySlotUnlessKept(releaseCtx, input.Slot, keep)
	}()
	if err := input.Action(ctx); err != nil {
		return err
	}
	keep = true
	return nil
}
