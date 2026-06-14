package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type RestoreLeaseInput struct {
	Limiter            ProviderAccountConcurrencyLimiter
	DataPlane          DataPlaneApplier
	Factory            SessionProviderFactory
	Lease              *proxyruntimev1.ProxyDynamicLease
	ProviderConfig     accountproxy.Config
	ProviderAccountID  string
	Limit              uint32
	DefaultTTL         time.Duration
	TTLBuffer          time.Duration
	SlotReleaseTimeout time.Duration
	LocalProtocol      string
	ResolveGateways    ProviderSessionGatewaysResolver
	ResolveLineBinding RouteLineBindingResolver
}

func RestoreLease(ctx context.Context, input RestoreLeaseInput) error {
	slot, err := acquireRestoreLeaseSlot(ctx, input)
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
				ProviderConfig:  input.ProviderConfig,
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

func acquireRestoreLeaseSlot(ctx context.Context, input RestoreLeaseInput) (ProviderAccountConcurrencySlot, error) {
	policy := ConcurrencyPolicy(input.Lease)
	return AcquireProviderAccountConcurrencySlot(
		ctx,
		input.Limiter,
		input.ProviderAccountID,
		input.Limit,
		policy,
		ConcurrencyHolder(input.Lease),
		ConcurrencySlotTTL(policy, input.DefaultTTL, input.TTLBuffer),
	)
}
