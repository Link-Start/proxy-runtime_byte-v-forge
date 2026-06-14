package lease

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CleanupFinalFailed   = "failed"
	CleanupFinalExpired  = "expired"
	CleanupFinalReleased = "released"
)

type FailedAcquireFactInput struct {
	LeaseID           string
	AccountID         string
	Purpose           string
	ProviderAccountID string
	Session           *proxyruntimev1.ProxySession
	Egress            *proxyruntimev1.ProxyEndpoint
	Listener          *proxyruntimev1.EgressListener
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	Message           string
	AcquiredAt        time.Time
}

type ActiveFactInput struct {
	LeaseID           string
	AccountID         string
	Purpose           string
	ProviderAccountID string
	Session           *proxyruntimev1.ProxySession
	Egress            *proxyruntimev1.ProxyEndpoint
	Listener          *proxyruntimev1.EgressListener
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func NewActiveFact(input ActiveFactInput) *proxyruntimev1.ProxyDynamicLease {
	lease := &proxyruntimev1.ProxyDynamicLease{
		LeaseId:           strings.TrimSpace(input.LeaseID),
		AccountId:         strings.TrimSpace(input.AccountID),
		Purpose:           defaultLeasePurpose(input.Purpose),
		ProviderAccountId: strings.TrimSpace(input.ProviderAccountID),
		Status:            proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE,
		Session:           input.Session,
		Egress:            input.Egress,
		Listener:          input.Listener,
		AcquiredAt:        timestamppb.New(input.AcquiredAt.UTC()),
		SelectionPlan:     input.SelectionPlan,
	}
	if input.Session != nil {
		lease.ExpiresAt = input.Session.GetExpiresAt()
	}
	return lease
}

func NewFailedAcquireFact(input FailedAcquireFactInput) *proxyruntimev1.ProxyDynamicLease {
	lease := &proxyruntimev1.ProxyDynamicLease{
		LeaseId:           strings.TrimSpace(input.LeaseID),
		AccountId:         strings.TrimSpace(input.AccountID),
		Purpose:           defaultLeasePurpose(input.Purpose),
		ProviderAccountId: strings.TrimSpace(input.ProviderAccountID),
		Status:            proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED,
		Session:           input.Session,
		Egress:            input.Egress,
		Listener:          input.Listener,
		AcquiredAt:        timestamppb.New(input.AcquiredAt.UTC()),
		SelectionPlan:     input.SelectionPlan,
		ErrorMessage:      defaultFailureMessage(input.Message),
	}
	if input.Session != nil {
		lease.ExpiresAt = input.Session.GetExpiresAt()
	}
	return lease
}

func MarkReleaseCleanupFailure(lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) {
	MarkCleanupPending(lease, routePending, providerPending, CleanupFinalReleased)
	MarkFailed(lease, message)
}

func MarkExpiredCleanupFailure(lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) {
	MarkCleanupPending(lease, routePending, providerPending, CleanupFinalExpired)
	MarkFailed(lease, message)
}

func MarkCleanupRetry(lease *proxyruntimev1.ProxyDynamicLease, message string) {
	if lease == nil {
		return
	}
	lease.ErrorMessage = strings.TrimSpace(message)
}

func MarkFailed(lease *proxyruntimev1.ProxyDynamicLease, message string) {
	if lease == nil {
		return
	}
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED
	lease.ErrorMessage = strings.TrimSpace(message)
}

func MarkExpired(lease *proxyruntimev1.ProxyDynamicLease) {
	if lease == nil {
		return
	}
	ClearCleanupPending(lease, true, true)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED
	lease.ErrorMessage = ""
}

func MarkReleased(lease *proxyruntimev1.ProxyDynamicLease) {
	if lease == nil {
		return
	}
	ClearCleanupPending(lease, true, true)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED
	lease.ErrorMessage = ""
}

func defaultLeasePurpose(value string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return "general"
}

func defaultFailureMessage(value string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return "lease acquire failed"
}
