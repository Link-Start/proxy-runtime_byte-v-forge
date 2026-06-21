package lease

import (
	"errors"
	"strconv"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

var (
	ErrAcquireRequestRequired   = errors.New("acquire request is required")
	ErrAcquireAccountIDRequired = errors.New("account_id is required")
	ErrReleaseRequestRequired   = errors.New("release request is required")
	ErrReleaseLookupRequired    = errors.New("lease_id or account_id is required")
	ErrReleaseAccountIDMismatch = errors.New("lease account_id mismatch")
	ErrReleasePurposeMismatch   = errors.New("lease purpose mismatch")
)

type ReleaseLookup struct {
	LeaseID   string
	AccountID string
	Purpose   string
}

func PrepareAcquireRequest(req *proxygatewayv1.AcquireProxyLeaseRequest) error {
	if req == nil {
		return ErrAcquireRequestRequired
	}
	req.AccountId = strings.TrimSpace(req.GetAccountId())
	if req.AccountId == "" {
		return ErrAcquireAccountIDRequired
	}
	req.Purpose = defaultLeasePurpose(req.GetPurpose())
	return nil
}

func ParseReleaseRequest(req *proxygatewayv1.ReleaseProxyLeaseRequest) (ReleaseLookup, error) {
	if req == nil {
		return ReleaseLookup{}, ErrReleaseRequestRequired
	}
	lookup := ReleaseLookup{
		LeaseID:   strings.TrimSpace(req.GetLeaseId()),
		AccountID: strings.TrimSpace(req.GetAccountId()),
		Purpose:   strings.TrimSpace(req.GetPurpose()),
	}
	if lookup.LeaseID == "" && lookup.AccountID == "" {
		return ReleaseLookup{}, ErrReleaseLookupRequired
	}
	return lookup, nil
}

func ValidateReleaseLeaseMatch(lookup ReleaseLookup, lease *proxygatewayv1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	if lookup.AccountID != "" && lookup.AccountID != lease.GetAccountId() {
		return ErrReleaseAccountIDMismatch
	}
	if lookup.Purpose != "" && lookup.Purpose != lease.GetPurpose() {
		return ErrReleasePurposeMismatch
	}
	return nil
}

func RequestedSessionID(req *proxygatewayv1.AcquireProxyLeaseRequest) string {
	labels := req.GetPolicy().GetLabels()
	return appcore.FirstNonEmpty(
		labels["session_id"],
		labels["sticky_session_id"],
		labels["sticky_id"],
		labels["sid"],
		labels["session"],
	)
}

func ApplyRequestLabels(req *proxygatewayv1.AcquireProxyLeaseRequest) {
	if req == nil {
		return
	}
	if req.Policy == nil {
		req.Policy = &proxygatewayv1.ProxySessionPolicy{}
	}
	if req.Policy.Labels == nil {
		req.Policy.Labels = map[string]string{}
	}
	req.Policy.Labels[LabelAccountID] = req.GetAccountId()
	req.Policy.Labels[LabelPurpose] = req.GetPurpose()
}

func SetAttemptLabel(req *proxygatewayv1.AcquireProxyLeaseRequest, attempt int) {
	if req == nil {
		return
	}
	ApplyRequestLabels(req)
	req.Policy.Labels[LabelAttempt] = strconv.Itoa(attempt)
}

func ApplyProviderSessionRequestLabels(req *proxygatewayv1.AcquireProxyLeaseRequest, selectionPlan *proxygatewayv1.ProxyDynamicIPSelectionPlan, concurrencyHolder string) {
	if req == nil {
		return
	}
	requestedSessionID := RequestedSessionID(req)
	ApplyRequestLabels(req)
	if requestedSessionID != "" {
		req.Policy.Labels[LabelSessionID] = requestedSessionID
	}
	req.Policy.Labels[LabelSelectionID] = selectionPlan.GetSelectionId()
	req.Policy.Labels[LabelDynamicIPEndpointID] = selectionPlan.GetSelectedEndpoint().GetEndpointId()
	req.Policy.Labels[LabelProviderAccountConcurrencyHolder] = concurrencyHolder
}
