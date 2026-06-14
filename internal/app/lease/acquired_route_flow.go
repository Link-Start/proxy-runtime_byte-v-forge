package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
	return ApplyAcquiredEndpointRoute(ctx, AcquiredEndpointRouteApplyInput{
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
}
