package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type AcquiredRouteApplyStage int

const (
	AcquiredRouteApplyNoError AcquiredRouteApplyStage = iota
	AcquiredRouteApplyDataPlane
	AcquiredRouteApplyFactSave
)

type AcquiredRouteApplyInput struct {
	Store             OrchestrationStore
	DataPlane         DataPlaneApplier
	Failure           *FailedAcquireRecorder
	Route             SessionRoute
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	ProviderAccountID string
	Session           *proxyruntimev1.ProxySession
	Egress            *proxyruntimev1.ProxyEndpoint
	Listener          *proxyruntimev1.EgressListener
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func ApplyAcquiredRoute(ctx context.Context, input AcquiredRouteApplyInput) (*proxyruntimev1.ProxyDynamicLease, AcquiredRouteApplyStage, error) {
	if err := UpsertSessionRoute(ctx, input.DataPlane, input.Route); err != nil {
		input.Failure.AfterRoute(ctx, input.Route, "dataplane route apply failed")
		return nil, AcquiredRouteApplyDataPlane, err
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
		input.Failure.AfterRoute(ctx, input.Route, "lease fact save failed")
		return nil, AcquiredRouteApplyFactSave, err
	}
	return lease, AcquiredRouteApplyNoError, nil
}
