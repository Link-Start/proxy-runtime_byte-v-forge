package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type AcquiredRouteFlowInput struct {
	Store             OrchestrationStore
	DataPlane         DataPlaneApplier
	Failure           *FailedAcquireRecorder
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxyruntimev1.ProxySession
	Nodes             []provider.Node
	DialerProxy       string
	LineLabels        map[string]string
	LocalProtocol     string
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
	Managed           bool
	FallbackProtocol  string
	ResolveListener   AcquiredEndpointListenerResolver
	ResolveEgress     AcquiredEndpointEgressResolver
	AfterApply        AcquiredRouteSuccessObserver
}

type AcquiredRouteApplier struct {
	Store            OrchestrationStore
	DataPlane        DataPlaneApplier
	Clock            clock.Clock
	LocalProtocol    string
	Managed          bool
	FallbackProtocol string
	ResolveListener  AcquiredEndpointListenerResolver
	ResolveEgress    AcquiredEndpointEgressResolver
	AfterApply       AcquiredRouteSuccessObserver
}

type AcquiredRouteApplierInput struct {
	Failure           *FailedAcquireRecorder
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxyruntimev1.ProxySession
	Nodes             []provider.Node
	DialerProxy       string
	LineLabels        map[string]string
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
}

type AcquiredRouteSuccessObserver func(context.Context, *proxyruntimev1.ProxyDynamicLease)

func (a AcquiredRouteApplier) Apply(ctx context.Context, input AcquiredRouteApplierInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	return ApplyAcquiredRouteFlow(ctx, AcquiredRouteFlowInput{
		Store:             a.Store,
		DataPlane:         a.DataPlane,
		Failure:           input.Failure,
		LeaseID:           input.LeaseID,
		Request:           input.Request,
		ProviderClient:    input.ProviderClient,
		ProviderAccountID: input.ProviderAccountID,
		ConcurrencyHolder: input.ConcurrencyHolder,
		Session:           input.Session,
		Nodes:             input.Nodes,
		DialerProxy:       input.DialerProxy,
		LineLabels:        input.LineLabels,
		LocalProtocol:     a.LocalProtocol,
		SelectionPlan:     input.SelectionPlan,
		AcquiredAt:        a.now(),
		Managed:           a.Managed,
		FallbackProtocol:  a.FallbackProtocol,
		ResolveListener:   a.ResolveListener,
		ResolveEgress:     a.ResolveEgress,
		AfterApply:        a.AfterApply,
	})
}

func (a AcquiredRouteApplier) now() time.Time {
	if a.Clock != nil {
		return a.Clock.Now()
	}
	return time.Now()
}

func ApplyAcquiredRouteFlow(ctx context.Context, input AcquiredRouteFlowInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	endpoint, err := MaterializeAcquiredEndpoint(ctx, AcquiredEndpointMaterializeInput{
		Request:           input.Request,
		LeaseID:           input.LeaseID,
		SelectionPlan:     input.SelectionPlan,
		ProviderClient:    input.ProviderClient,
		ProviderAccountID: input.ProviderAccountID,
		ConcurrencyHolder: input.ConcurrencyHolder,
		Session:           input.Session,
		LineLabels:        input.LineLabels,
		Failure:           input.Failure,
		Managed:           input.Managed,
		FallbackProtocol:  input.FallbackProtocol,
		ResolveListener:   input.ResolveListener,
		ResolveEgress:     input.ResolveEgress,
	})
	if err != nil {
		return nil, err
	}
	lease, err := ApplyAcquiredEndpointRoute(ctx, AcquiredEndpointRouteApplyInput{
		Store:             input.Store,
		DataPlane:         input.DataPlane,
		Failure:           input.Failure,
		LeaseID:           input.LeaseID,
		Request:           input.Request,
		ProviderAccountID: input.ProviderAccountID,
		Session:           input.Session,
		Egress:            endpoint.Egress,
		Listener:          endpoint.Listener,
		ListenerProto:     endpoint.ListenerProto,
		Nodes:             input.Nodes,
		DialerProxy:       input.DialerProxy,
		LocalProtocol:     input.LocalProtocol,
		SelectionPlan:     input.SelectionPlan,
		AcquiredAt:        input.AcquiredAt,
	})
	if err != nil {
		return nil, err
	}
	if input.AfterApply != nil {
		input.AfterApply(ctx, lease)
	}
	return lease, nil
}
