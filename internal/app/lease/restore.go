package lease

import (
	"context"
	"errors"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

var ErrLeaseRouteRestorerRequired = errors.New("lease route restorer is required")

type LeaseRouteRestorerResolver func(context.Context, *proxygatewayv1.ProxyDynamicLease) (LeaseRouteRestorer, error)

type RestoreLeaseRouteRunner struct {
	ResolveRestorer LeaseRouteRestorerResolver
}

func (r RestoreLeaseRouteRunner) Restore(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	if r.ResolveRestorer == nil {
		return ErrLeaseRouteRestorerRequired
	}
	restorer, err := r.ResolveRestorer(ctx, lease)
	if err != nil {
		return err
	}
	return restorer.Restore(ctx, lease)
}

var ErrRestoreLeaseRouteRequired = errors.New("lease session or listener is missing")

type RestoreLeaseLimitFunc func(*proxygatewayv1.ProxyDynamicLease) uint32

type LeaseRouteRestorer struct {
	Limiter            ProviderAccountConcurrencyLimiter
	Store              OrchestrationStore
	DataPlane          DataPlaneApplier
	Factory            SessionProviderFactory
	Limit              RestoreLeaseLimitFunc
	DefaultTTL         time.Duration
	TTLBuffer          time.Duration
	SlotReleaseTimeout time.Duration
	LocalProtocol      string
	ResolveGateways    ProviderSessionGatewaysResolverFactory
	ResolveLineBinding RouteLineBindingResolver
}

type RestoreLeaseInput struct {
	Limiter            ProviderAccountConcurrencyLimiter
	Store              OrchestrationStore
	DataPlane          DataPlaneApplier
	Factory            SessionProviderFactory
	Lease              *proxygatewayv1.ProxyDynamicLease
	Limit              uint32
	DefaultTTL         time.Duration
	TTLBuffer          time.Duration
	SlotReleaseTimeout time.Duration
	LocalProtocol      string
	ResolveGateways    ProviderSessionGatewaysResolver
	ResolveLineBinding RouteLineBindingResolver
}

func (r LeaseRouteRestorer) Restore(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	return RestoreLease(ctx, RestoreLeaseInput{
		Limiter:            r.Limiter,
		Store:              r.Store,
		DataPlane:          r.DataPlane,
		Factory:            r.Factory,
		Lease:              lease,
		Limit:              r.limit(lease),
		DefaultTTL:         r.DefaultTTL,
		TTLBuffer:          r.TTLBuffer,
		SlotReleaseTimeout: r.SlotReleaseTimeout,
		LocalProtocol:      r.LocalProtocol,
		ResolveGateways:    r.resolveGateways(lease),
		ResolveLineBinding: r.ResolveLineBinding,
	})
}

func (r LeaseRouteRestorer) limit(lease *proxygatewayv1.ProxyDynamicLease) uint32 {
	if r.Limit == nil {
		return 0
	}
	return r.Limit(lease)
}

func (r LeaseRouteRestorer) resolveGateways(lease *proxygatewayv1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	if r.ResolveGateways == nil {
		return nil
	}
	return r.ResolveGateways(lease)
}

func RestoreLease(ctx context.Context, input RestoreLeaseInput) error {
	if input.Lease.GetSession() == nil || input.Lease.GetListener() == nil {
		return ErrRestoreLeaseRouteRequired
	}
	providerCfg, providerAccountID, err := ProviderConfigForLease(ctx, input.Store, input.Lease)
	if err != nil {
		return err
	}
	slot, err := acquireRestoreLeaseSlot(ctx, input, providerAccountID)
	if err != nil {
		return err
	}
	return RunTemporaryConcurrencySlot(ctx, TemporaryConcurrencySlotInput{
		Slot:           slot,
		ReleaseTimeout: input.SlotReleaseTimeout,
		Action: func(ctx context.Context) error {
			nodes, err := FetchLeaseProviderSession(ctx, ProviderSessionFetchInput{
				Factory:         input.Factory,
				Lease:           input.Lease,
				ProviderConfig:  providerCfg,
				ResolveGateways: input.ResolveGateways,
			})
			if err != nil {
				return err
			}
			return RestoreLeaseRoute(ctx, RestoreRouteInput{
				DataPlane:          input.DataPlane,
				Lease:              input.Lease,
				Nodes:              nodes,
				LocalProtocol:      input.LocalProtocol,
				ResolveLineBinding: input.ResolveLineBinding,
			})
		},
	})
}

func acquireRestoreLeaseSlot(ctx context.Context, input RestoreLeaseInput, providerAccountID string) (ProviderAccountConcurrencySlot, error) {
	policy := ConcurrencyPolicy(input.Lease)
	return AcquireProviderAccountConcurrencySlot(
		ctx,
		input.Limiter,
		providerAccountID,
		input.Limit,
		policy,
		ConcurrencyHolder(input.Lease),
		ConcurrencySlotTTL(policy, input.DefaultTTL, input.TTLBuffer),
	)
}

type RestoreRouteInput struct {
	DataPlane          DataPlaneApplier
	Lease              *proxygatewayv1.ProxyDynamicLease
	Nodes              []provider.Node
	LocalProtocol      string
	ResolveLineBinding RouteLineBindingResolver
}

func RestoreLeaseRoute(ctx context.Context, input RestoreRouteInput) error {
	lineBinding, err := PrepareRouteLineBinding(ctx, input.Nodes, input.Lease.GetAccountId(), input.ResolveLineBinding)
	if err != nil {
		return err
	}
	route, ok := SessionRouteFromLease(input.Lease, lineBinding.Nodes, lineBinding.DialerProxy, input.LocalProtocol)
	if !ok {
		return nil
	}
	return UpsertSessionRoute(ctx, input.DataPlane, route)
}
