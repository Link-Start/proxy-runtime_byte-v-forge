package lease

import (
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
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
	Session           *proxygatewayv1.ProxySession
	Egress            *proxygatewayv1.ProxyEndpoint
	Listener          *proxygatewayv1.EgressListener
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	Message           string
	AcquiredAt        time.Time
}

type ActiveFactInput struct {
	LeaseID           string
	AccountID         string
	Purpose           string
	ProviderAccountID string
	Session           *proxygatewayv1.ProxySession
	Egress            *proxygatewayv1.ProxyEndpoint
	Listener          *proxygatewayv1.EgressListener
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	AcquiredAt        time.Time
}

func NewActiveFact(input ActiveFactInput) *proxygatewayv1.ProxyDynamicLease {
	lease := &proxygatewayv1.ProxyDynamicLease{
		LeaseId:           strings.TrimSpace(input.LeaseID),
		AccountId:         strings.TrimSpace(input.AccountID),
		Purpose:           defaultLeasePurpose(input.Purpose),
		ProviderAccountId: strings.TrimSpace(input.ProviderAccountID),
		Status:            proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE,
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

func NewFailedAcquireFact(input FailedAcquireFactInput) *proxygatewayv1.ProxyDynamicLease {
	lease := &proxygatewayv1.ProxyDynamicLease{
		LeaseId:           strings.TrimSpace(input.LeaseID),
		AccountId:         strings.TrimSpace(input.AccountID),
		Purpose:           defaultLeasePurpose(input.Purpose),
		ProviderAccountId: strings.TrimSpace(input.ProviderAccountID),
		Status:            proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED,
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

func MarkReleaseCleanupFailure(lease *proxygatewayv1.ProxyDynamicLease, routePending bool, providerPending bool, message string) {
	MarkCleanupPending(lease, routePending, providerPending, CleanupFinalReleased)
	MarkFailed(lease, message)
}

func MarkExpiredCleanupFailure(lease *proxygatewayv1.ProxyDynamicLease, routePending bool, providerPending bool, message string) {
	MarkCleanupPending(lease, routePending, providerPending, CleanupFinalExpired)
	MarkFailed(lease, message)
}

func MarkCleanupRetry(lease *proxygatewayv1.ProxyDynamicLease, message string) {
	if lease == nil {
		return
	}
	lease.ErrorMessage = strings.TrimSpace(message)
}

func MarkFailed(lease *proxygatewayv1.ProxyDynamicLease, message string) {
	if lease == nil {
		return
	}
	lease.Status = proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED
	lease.ErrorMessage = strings.TrimSpace(message)
}

func MarkExpired(lease *proxygatewayv1.ProxyDynamicLease) {
	if lease == nil {
		return
	}
	ClearCleanupPending(lease, true, true)
	lease.Status = proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED
	lease.ErrorMessage = ""
}

func MarkReleased(lease *proxygatewayv1.ProxyDynamicLease) {
	if lease == nil {
		return
	}
	ClearCleanupPending(lease, true, true)
	lease.Status = proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED
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
