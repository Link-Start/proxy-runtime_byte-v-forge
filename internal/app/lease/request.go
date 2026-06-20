package lease

import (
	"errors"
	"strconv"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
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

func PrepareAcquireRequest(req *proxyruntimev1.AcquireProxyLeaseRequest) error {
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

func ParseReleaseRequest(req *proxyruntimev1.ReleaseProxyLeaseRequest) (ReleaseLookup, error) {
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

func ValidateReleaseLeaseMatch(lookup ReleaseLookup, lease *proxyruntimev1.ProxyDynamicLease) error {
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

func RequestedSessionID(req *proxyruntimev1.AcquireProxyLeaseRequest) string {
	labels := req.GetPolicy().GetLabels()
	return appcore.FirstNonEmpty(
		labels["session_id"],
		labels["sticky_session_id"],
		labels["sticky_id"],
		labels["sid"],
		labels["session"],
	)
}

func ApplyRequestLabels(req *proxyruntimev1.AcquireProxyLeaseRequest) {
	if req == nil {
		return
	}
	if req.Policy == nil {
		req.Policy = &proxyruntimev1.ProxySessionPolicy{}
	}
	if req.Policy.Labels == nil {
		req.Policy.Labels = map[string]string{}
	}
	req.Policy.Labels[LabelAccountID] = req.GetAccountId()
	req.Policy.Labels[LabelPurpose] = req.GetPurpose()
}

func SetAttemptLabel(req *proxyruntimev1.AcquireProxyLeaseRequest, attempt int) {
	if req == nil {
		return
	}
	ApplyRequestLabels(req)
	req.Policy.Labels[LabelAttempt] = strconv.Itoa(attempt)
}

func ApplyProviderSessionRequestLabels(req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, concurrencyHolder string) {
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
