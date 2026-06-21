package lease

import (
	"context"
	"errors"
	"fmt"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

type AcquiredSessionRouteInput struct {
	Session       *proxygatewayv1.ProxySession
	Egress        *proxygatewayv1.ProxyEndpoint
	Listener      Listener
	Nodes         []provider.Node
	DialerProxy   string
	LocalProtocol string
}

func NewAcquiredSessionRoute(input AcquiredSessionRouteInput) SessionRoute {
	if input.Session != nil {
		input.Session.Egress = input.Egress
	}
	return NewSessionRoute(input.Session.GetSessionId(), input.Listener, input.Nodes, input.DialerProxy, input.LocalProtocol)
}

var (
	ErrAcquiredRouteDataPlane = errors.New("dataplane route apply failed")
	ErrAcquiredRouteFactSave  = errors.New("lease fact save failed")
)

type AcquiredRouteApplyInput struct {
	Store             OrchestrationStore
	DataPlane         DataPlaneApplier
	Failure           *FailedAcquireRecorder
	Route             SessionRoute
	LeaseID           string
	Request           *proxygatewayv1.AcquireProxyLeaseRequest
	ProviderAccountID string
	Session           *proxygatewayv1.ProxySession
	Egress            *proxygatewayv1.ProxyEndpoint
	Listener          *proxygatewayv1.EgressListener
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func ApplyAcquiredRoute(ctx context.Context, input AcquiredRouteApplyInput) (*proxygatewayv1.ProxyDynamicLease, error) {
	if err := UpsertSessionRoute(ctx, input.DataPlane, input.Route); err != nil {
		input.Failure.AfterRoute(ctx, input.Route, ErrAcquiredRouteDataPlane.Error())
		return nil, fmt.Errorf("%w: %w", ErrAcquiredRouteDataPlane, err)
	}
	lease, err := SaveAcquiredActiveFact(ctx, input.Store, AcquiredActiveFactInput{
		LeaseID:           input.LeaseID,
		Request:           input.Request,
		ProviderAccountID: input.ProviderAccountID,
		Session:           input.Session,
		Egress:            input.Egress,
		Listener:          input.Listener,
		SelectionPlan:     input.SelectionPlan,
		AcquiredAt:        input.AcquiredAt,
	})
	if err != nil {
		input.Failure.AfterRoute(ctx, input.Route, ErrAcquiredRouteFactSave.Error())
		return nil, fmt.Errorf("%w: %w", ErrAcquiredRouteFactSave, err)
	}
	return lease, nil
}

type AcquiredRouteFlowInput struct {
	Store             OrchestrationStore
	DataPlane         DataPlaneApplier
	Failure           *FailedAcquireRecorder
	LeaseID           string
	Request           *proxygatewayv1.AcquireProxyLeaseRequest
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxygatewayv1.ProxySession
	Nodes             []provider.Node
	DialerProxy       string
	LineLabels        map[string]string
	LocalProtocol     string
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
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
	Request           *proxygatewayv1.AcquireProxyLeaseRequest
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxygatewayv1.ProxySession
	Nodes             []provider.Node
	DialerProxy       string
	LineLabels        map[string]string
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
}

type AcquiredRouteSuccessObserver func(context.Context, *proxygatewayv1.ProxyDynamicLease)

func (a AcquiredRouteApplier) Apply(ctx context.Context, input AcquiredRouteApplierInput) (*proxygatewayv1.ProxyDynamicLease, error) {
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

func ApplyAcquiredRouteFlow(ctx context.Context, input AcquiredRouteFlowInput) (*proxygatewayv1.ProxyDynamicLease, error) {
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
