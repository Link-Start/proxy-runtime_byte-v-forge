package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type RestoreLeaseInput struct {
	Limiter            ProviderAccountConcurrencyLimiter
	Store              OrchestrationStore
	DataPlane          DataPlaneApplier
	Factory            SessionProviderFactory
	Lease              *proxyruntimev1.ProxyDynamicLease
	Limit              uint32
	DefaultTTL         time.Duration
	TTLBuffer          time.Duration
	SlotReleaseTimeout time.Duration
	LocalProtocol      string
	ResolveGateways    ProviderSessionGatewaysResolver
	ResolveLineBinding RouteLineBindingResolver
}

func RestoreLease(ctx context.Context, input RestoreLeaseInput) error {
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
