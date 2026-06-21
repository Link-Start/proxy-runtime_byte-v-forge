package store

import (
	"errors"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"
)

type DynamicLeaseFactSave struct {
	LeaseID           string
	AccountID         string
	Purpose           string
	ProviderAccountID string
	Status            proxygatewayv1.ProxyDynamicLeaseStatus
	JSON              string
}

func PrepareDynamicLeaseFactSave(lease *proxygatewayv1.ProxyDynamicLease) (DynamicLeaseFactSave, error) {
	if lease == nil || strings.TrimSpace(lease.GetLeaseId()) == "" {
		return DynamicLeaseFactSave{}, errors.New("lease_id is required")
	}
	if strings.TrimSpace(lease.GetAccountId()) == "" {
		return DynamicLeaseFactSave{}, errors.New("lease account_id is required")
	}
	status := lease.GetStatus()
	if status == proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_UNSPECIFIED {
		status = proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE
		lease.Status = status
	}
	data, err := protojsoncodec.Marshal(lease)
	if err != nil {
		return DynamicLeaseFactSave{}, err
	}
	return DynamicLeaseFactSave{
		LeaseID:           strings.TrimSpace(lease.GetLeaseId()),
		AccountID:         strings.TrimSpace(lease.GetAccountId()),
		Purpose:           strings.TrimSpace(lease.GetPurpose()),
		ProviderAccountID: strings.TrimSpace(lease.GetProviderAccountId()),
		Status:            status,
		JSON:              string(data),
	}, nil
}
