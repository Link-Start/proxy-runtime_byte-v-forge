package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type AcquiredActiveFactInput struct {
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	ProviderAccountID string
	Session           *proxyruntimev1.ProxySession
	Egress            *proxyruntimev1.ProxyEndpoint
	Listener          *proxyruntimev1.EgressListener
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func SaveAcquiredActiveFact(ctx context.Context, store OrchestrationStore, input AcquiredActiveFactInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	return SaveActiveFact(ctx, store, ActiveFactInput{
		LeaseID:           input.LeaseID,
		AccountID:         input.Request.GetAccountId(),
		Purpose:           input.Request.GetPurpose(),
		ProviderAccountID: input.ProviderAccountID,
		Session:           input.Session,
		Egress:            input.Egress,
		Listener:          input.Listener,
		SelectionPlan:     input.SelectionPlan,
		AcquiredAt:        input.AcquiredAt,
	})
}
