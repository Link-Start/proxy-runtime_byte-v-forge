package lease

import (
	"context"
	"errors"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

type ExistingActiveLeaseDecision int

const (
	ExistingActiveLeaseIgnore ExistingActiveLeaseDecision = iota
	ExistingActiveLeaseReuse
	ExistingActiveLeaseReplace
)

type ExistingActiveLeaseAction func(context.Context, *proxygatewayv1.ProxyDynamicLease) error

type ExistingActiveLeaseInput struct {
	Store               OrchestrationStore
	Request             *proxygatewayv1.AcquireProxyLeaseRequest
	Now                 time.Time
	PlaygroundAccountID string
	PlaygroundUsername  string
	Reuse               ExistingActiveLeaseAction
	Replace             ExistingActiveLeaseAction
}

func HandleExistingActiveLease(ctx context.Context, input ExistingActiveLeaseInput) (*proxygatewayv1.ProxyDynamicLease, bool, error) {
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

func DecideExistingActiveLease(req *proxygatewayv1.AcquireProxyLeaseRequest, lease *proxygatewayv1.ProxyDynamicLease, now time.Time, playgroundAccountID string, playgroundUsername string) ExistingActiveLeaseDecision {
	if !ActiveAt(lease, now) {
		return ExistingActiveLeaseIgnore
	}
	if !req.GetForceNew() && !PlaygroundLeaseNeedsReplacement(req, lease, playgroundAccountID, playgroundUsername) {
		return ExistingActiveLeaseReuse
	}
	return ExistingActiveLeaseReplace
}

func runExistingActiveLeaseAction(ctx context.Context, action ExistingActiveLeaseAction, lease *proxygatewayv1.ProxyDynamicLease) error {
	if action == nil {
		return nil
	}
	return action(ctx, lease)
}

func ActiveLeaseByRequest(ctx context.Context, store OrchestrationStore, req *proxygatewayv1.AcquireProxyLeaseRequest, sessionID string) (*proxygatewayv1.ProxyDynamicLease, error) {
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
	Request           *proxygatewayv1.AcquireProxyLeaseRequest
	ProviderAccountID string
	Session           *proxygatewayv1.ProxySession
	Egress            *proxygatewayv1.ProxyEndpoint
	Listener          *proxygatewayv1.EgressListener
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func SaveAcquiredActiveFact(ctx context.Context, store OrchestrationStore, input AcquiredActiveFactInput) (*proxygatewayv1.ProxyDynamicLease, error) {
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

func ApplyAcquireRequestPolicies(req *proxygatewayv1.AcquireProxyLeaseRequest, profiles []*proxygatewayv1.EgressProfileSettings) (*proxygatewayv1.ProxyDynamicIPSelectionPolicy, error) {
	req.Policy = kernel.NormalizeDynamicIPSessionPolicy(req.GetPolicy())
	ApplyRequestLabels(req)
	if err := ApplyProfileDynamicIPPolicy(profiles, req); err != nil {
		return nil, err
	}
	return NormalizeDynamicIPSelectionPolicy(req), nil
}
