package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *RuntimeService) GetProxyExitIP(ctx context.Context, req *proxyruntimev1.GetProxyExitIPRequest) (*proxyruntimev1.GetProxyExitIPResponse, error) {
	return s.checks.GetProxyExitIP(ctx, req)
}

func (s *RuntimeService) GetProxyExitGeo(ctx context.Context, req *proxyruntimev1.GetProxyExitGeoRequest) (*proxyruntimev1.GetProxyExitGeoResponse, error) {
	return s.checks.GetProxyExitGeo(ctx, req)
}

func (s *RuntimeService) CheckProxyIPFraud(ctx context.Context, req *proxyruntimev1.CheckProxyIPFraudRequest) (*proxyruntimev1.CheckProxyIPFraudResponse, error) {
	return s.checks.CheckProxyIPFraud(ctx, req)
}

func (s *RuntimeService) CheckProxyEdgeAccess(ctx context.Context, req *proxyruntimev1.CheckProxyEdgeAccessRequest) (*proxyruntimev1.CheckProxyEdgeAccessResponse, error) {
	return s.checks.CheckProxyEdgeAccess(ctx, req)
}

func (s *RuntimeService) CheckProxyTargetConnectivity(ctx context.Context, req *proxyruntimev1.CheckProxyTargetConnectivityRequest) (*proxyruntimev1.CheckProxyTargetConnectivityResponse, error) {
	return s.checks.CheckProxyTargetConnectivity(ctx, req)
}

func (s *RuntimeService) GetProxyExitCheckSnapshot(ctx context.Context, req *proxyruntimev1.GetProxyExitCheckSnapshotRequest) (*proxyruntimev1.GetProxyExitCheckSnapshotResponse, error) {
	return s.checks.GetProxyExitCheckSnapshot(ctx, req)
}
