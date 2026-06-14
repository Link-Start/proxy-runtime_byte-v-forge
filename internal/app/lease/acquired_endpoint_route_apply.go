package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

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
