package lease

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

type ExistingActiveLeaseDecision int

const (
	ExistingActiveLeaseIgnore ExistingActiveLeaseDecision = iota
	ExistingActiveLeaseReuse
	ExistingActiveLeaseReplace
)

type ExistingActiveLeaseAction func(context.Context, *proxyruntimev1.ProxyDynamicLease) error

type ExistingActiveLeaseInput struct {
	Store               OrchestrationStore
	Request             *proxyruntimev1.AcquireProxyLeaseRequest
	Now                 time.Time
	PlaygroundAccountID string
	PlaygroundUsername  string
	Reuse               ExistingActiveLeaseAction
	Replace             ExistingActiveLeaseAction
}

func HandleExistingActiveLease(ctx context.Context, input ExistingActiveLeaseInput) (*proxyruntimev1.ProxyDynamicLease, bool, error) {
	existing, err := ActiveLeaseByRequest(ctx, input.Store, input.Request, RequestedSessionID(input.Request))
	if err != nil {
		return nil, false, nil
	}
	switch DecideExistingActiveLease(input.Request, existing, input.Now, input.PlaygroundAccountID, input.PlaygroundUsername) {
	case ExistingActiveLeaseReuse:
		return existing, true, runExistingActiveLeaseAction(ctx, input.Reuse, existing)
	case ExistingActiveLeaseReplace:
		return nil, false, runExistingActiveLeaseAction(ctx, input.Replace, existing)
	default:
		return nil, false, nil
	}
}

func DecideExistingActiveLease(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease, now time.Time, playgroundAccountID string, playgroundUsername string) ExistingActiveLeaseDecision {
	if !ActiveAt(lease, now) {
		return ExistingActiveLeaseIgnore
	}
	if !req.GetForceNew() && !PlaygroundLeaseNeedsReplacement(req, lease, playgroundAccountID, playgroundUsername) {
		return ExistingActiveLeaseReuse
	}
	return ExistingActiveLeaseReplace
}

func runExistingActiveLeaseAction(ctx context.Context, action ExistingActiveLeaseAction, lease *proxyruntimev1.ProxyDynamicLease) error {
	if action == nil {
		return nil
	}
	return action(ctx, lease)
}

func ActiveLeaseByRequest(ctx context.Context, store OrchestrationStore, req *proxyruntimev1.AcquireProxyLeaseRequest, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if store == nil {
		return nil, errors.New("lease store is required")
	}
	if sessionID != "" {
		return store.ActiveLeaseFactBySession(ctx, req.GetAccountId(), req.GetPurpose(), sessionID)
	}
	return store.ActiveLeaseFactByAccount(ctx, req.GetAccountId(), req.GetPurpose())
}

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

func ApplyAcquireRequestPolicies(req *proxyruntimev1.AcquireProxyLeaseRequest, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyDynamicIPSelectionPolicy, error) {
	req.Policy = kernel.NormalizeDynamicIPSessionPolicy(req.GetPolicy())
	ApplyRequestLabels(req)
	if err := ApplyProfileDynamicIPPolicy(profiles, req); err != nil {
		return nil, err
	}
	return NormalizeDynamicIPSelectionPolicy(req), nil
}
