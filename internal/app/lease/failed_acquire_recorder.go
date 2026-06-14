package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type FailedAcquireRecorderInput struct {
	Store             OrchestrationStore
	IDs               IDGenerator
	Clock             Clock
	DataPlane         DataPlaneApplier
	Logger            Logger
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	ProviderAccountID string
	ProviderClient    SessionProvider
	Session           *proxyruntimev1.ProxySession
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
}

type FailedAcquireRecorder struct {
	store             OrchestrationStore
	ids               IDGenerator
	clock             Clock
	dataPlane         DataPlaneApplier
	logger            Logger
	request           *proxyruntimev1.AcquireProxyLeaseRequest
	providerAccountID string
	providerClient    SessionProvider
	session           *proxyruntimev1.ProxySession
	listener          *proxyruntimev1.EgressListener
	egress            *proxyruntimev1.ProxyEndpoint
	selectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
}

func NewFailedAcquireRecorder(input FailedAcquireRecorderInput) *FailedAcquireRecorder {
	return &FailedAcquireRecorder{
		store:             input.Store,
		ids:               input.IDs,
		clock:             input.Clock,
		dataPlane:         input.DataPlane,
		logger:            input.Logger,
		request:           input.Request,
		providerAccountID: input.ProviderAccountID,
		providerClient:    input.ProviderClient,
		session:           input.Session,
		selectionPlan:     input.SelectionPlan,
	}
}

func (r *FailedAcquireRecorder) SetListener(listener *proxyruntimev1.EgressListener) {
	if r != nil {
		r.listener = listener
	}
}

func (r *FailedAcquireRecorder) SetEgress(egress *proxyruntimev1.ProxyEndpoint) {
	if r != nil {
		r.egress = egress
	}
}

func (r *FailedAcquireRecorder) SetEndpoint(endpoint AcquiredEndpoint) {
	if r == nil {
		return
	}
	r.listener = endpoint.ListenerProto
	r.egress = endpoint.Egress
}

func (r *FailedAcquireRecorder) BeforeRoute(ctx context.Context, message string) {
	if r == nil {
		return
	}
	r.warnProviderCleanup(MarkFailedAcquireBeforeRouteCleanup(ctx, r.providerClient, r.session))
	r.save(ctx, message)
}

func (r *FailedAcquireRecorder) AfterRoute(ctx context.Context, route SessionRoute, message string) {
	if r == nil {
		return
	}
	r.warnProviderCleanup(MarkFailedAcquireAfterRouteCleanup(ctx, r.dataPlane, route, r.providerClient, r.session))
	r.save(ctx, message)
}

func (r *FailedAcquireRecorder) warnProviderCleanup(err error) {
	if r == nil || err == nil || r.logger == nil {
		return
	}
	r.logger.Warn("provider session cleanup failed", "provider_id", ProviderName(r.providerClient), LabelAccountID, r.request.GetAccountId())
}

func (r *FailedAcquireRecorder) save(ctx context.Context, message string) {
	if r == nil {
		return
	}
	lease, err := SaveFailedAcquireFact(ctx, r.store, r.ids, r.clock, r.request, r.providerAccountID, r.session, r.egress, r.listener, r.selectionPlan, message)
	if err != nil && lease != nil && r.logger != nil {
		r.logger.Warn("save failed proxy lease fact failed", LabelAccountID, lease.GetAccountId(), LabelProviderAccountID, lease.GetProviderAccountId())
	}
}
