package lease

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

var (
	ErrAcquiredEndpointListenerResolverRequired = errors.New("acquired endpoint listener resolver is required")
	ErrAcquiredEndpointEgressResolverRequired   = errors.New("acquired endpoint egress resolver is required")
)

type AcquiredEndpointInput struct {
	Listener          Listener
	Egress            *proxyruntimev1.ProxyEndpoint
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxyruntimev1.ProxySession
	LineLabels        map[string]string
	Managed           bool
	FallbackProtocol  string
}

type AcquiredEndpoint struct {
	Listener      Listener
	ListenerProto *proxyruntimev1.EgressListener
	Egress        *proxyruntimev1.ProxyEndpoint
}

type AcquiredEndpointListenerResolver func(context.Context, string, string) (Listener, error)
type AcquiredEndpointEgressResolver func(context.Context, Listener) (*proxyruntimev1.ProxyEndpoint, error)

type AcquiredEndpointMaterializeInput struct {
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	LeaseID           string
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxyruntimev1.ProxySession
	LineLabels        map[string]string
	Failure           *FailedAcquireRecorder
	Managed           bool
	FallbackProtocol  string
	ResolveListener   AcquiredEndpointListenerResolver
	ResolveEgress     AcquiredEndpointEgressResolver
}

func MaterializeAcquiredEndpoint(ctx context.Context, input AcquiredEndpointMaterializeInput) (AcquiredEndpoint, error) {
	if input.ResolveListener == nil {
		return AcquiredEndpoint{}, ErrAcquiredEndpointListenerResolverRequired
	}
	if input.ResolveEgress == nil {
		return AcquiredEndpoint{}, ErrAcquiredEndpointEgressResolverRequired
	}
	listener, err := input.ResolveListener(ctx, input.Request.GetAccountId(), input.LeaseID)
	if err != nil {
		input.Failure.BeforeRoute(ctx, "lease listener allocation failed")
		return AcquiredEndpoint{}, err
	}
	egress, err := input.ResolveEgress(ctx, listener)
	if err != nil {
		input.Failure.BeforeRoute(ctx, "lease endpoint build failed")
		return AcquiredEndpoint{}, err
	}
	endpoint := NewAcquiredEndpoint(AcquiredEndpointInput{
		Listener:          listener,
		Egress:            egress,
		Request:           input.Request,
		SelectionPlan:     input.SelectionPlan,
		ProviderClient:    input.ProviderClient,
		ProviderAccountID: input.ProviderAccountID,
		ConcurrencyHolder: input.ConcurrencyHolder,
		Session:           input.Session,
		LineLabels:        input.LineLabels,
		Managed:           input.Managed,
		FallbackProtocol:  input.FallbackProtocol,
	})
	input.Failure.SetEndpoint(endpoint)
	return endpoint, nil
}

func NewAcquiredEndpoint(input AcquiredEndpointInput) AcquiredEndpoint {
	listenerProto := EgressListenerProto(input.Listener, input.Managed, input.FallbackProtocol)
	ApplyAcquiredEndpointMetadata(input.Egress, AcquiredEndpointMetadataInput{
		Request:           input.Request,
		SelectionPlan:     input.SelectionPlan,
		ProviderClient:    input.ProviderClient,
		ProviderAccountID: input.ProviderAccountID,
		ConcurrencyHolder: input.ConcurrencyHolder,
		Session:           input.Session,
		LineLabels:        input.LineLabels,
	})
	return AcquiredEndpoint{Listener: input.Listener, ListenerProto: listenerProto, Egress: input.Egress}
}

type AcquiredEndpointRouteApplyInput struct {
	Store             OrchestrationStore
	DataPlane         DataPlaneApplier
	Failure           *FailedAcquireRecorder
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	ProviderAccountID string
	Session           *proxyruntimev1.ProxySession
	Egress            *proxyruntimev1.ProxyEndpoint
	Listener          Listener
	ListenerProto     *proxyruntimev1.EgressListener
	Nodes             []provider.Node
	DialerProxy       string
	LocalProtocol     string
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func ApplyAcquiredEndpointRoute(ctx context.Context, input AcquiredEndpointRouteApplyInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	route := NewAcquiredSessionRoute(AcquiredSessionRouteInput{
		Session:       input.Session,
		Egress:        input.Egress,
		Listener:      input.Listener,
		Nodes:         input.Nodes,
		DialerProxy:   input.DialerProxy,
		LocalProtocol: input.LocalProtocol,
	})
	return ApplyAcquiredRoute(ctx, AcquiredRouteApplyInput{
		Store:             input.Store,
		DataPlane:         input.DataPlane,
		Failure:           input.Failure,
		Route:             route,
		LeaseID:           input.LeaseID,
		Request:           input.Request,
		ProviderAccountID: input.ProviderAccountID,
		Session:           input.Session,
		Egress:            input.Egress,
		Listener:          input.ListenerProto,
		SelectionPlan:     input.SelectionPlan,
		AcquiredAt:        input.AcquiredAt,
	})
}
