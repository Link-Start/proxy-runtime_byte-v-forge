package app

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type runtimeCheckApplication struct {
	runtime *Runtime
}

func newRuntimeCheckApplication(runtime *Runtime) runtimeCheckApplication {
	return runtimeCheckApplication{runtime: runtime}
}

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

func (a runtimeCheckApplication) GetProxyExitIP(ctx context.Context, req *proxyruntimev1.GetProxyExitIPRequest) (*proxyruntimev1.GetProxyExitIPResponse, error) {
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	timeout := proxyExitIPTimeout(settings)
	client, err := a.runtime.checkProxyHTTPClient(ctx, req.GetListenerId(), timeout)
	if err != nil {
		return nil, err
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ip, err := a.runtime.probeExitIP(probeCtx, client)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyExitIPResponse{ProxyExitIp: &proxyruntimev1.ProxyExitIP{Ip: ip, CheckedAt: timestamppb.Now()}}, nil
}

func (a runtimeCheckApplication) GetProxyExitGeo(ctx context.Context, req *proxyruntimev1.GetProxyExitGeoRequest) (*proxyruntimev1.GetProxyExitGeoResponse, error) {
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		return nil, errors.New("ip is required")
	}
	geo, err := a.runtime.lookupIPGeo(ctx, ip)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyExitGeoResponse{ProxyExitGeo: &proxyruntimev1.ProxyExitGeo{Ip: ip, CountryCode: geo.CountryCode, Region: geo.Region, City: geo.City, CheckedAt: timestamppb.Now()}}, nil
}

func (a runtimeCheckApplication) CheckProxyIPFraud(ctx context.Context, req *proxyruntimev1.CheckProxyIPFraudRequest) (*proxyruntimev1.CheckProxyIPFraudResponse, error) {
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		return nil, errors.New("ip is required")
	}
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	check, err := a.runtime.checkIPFraud(ctx, ip, settings)
	if err != nil {
		return nil, errors.New("check IP fraud")
	}
	return &proxyruntimev1.CheckProxyIPFraudResponse{Check: check}, nil
}

func (a runtimeCheckApplication) CheckProxyEdgeAccess(ctx context.Context, req *proxyruntimev1.CheckProxyEdgeAccessRequest) (*proxyruntimev1.CheckProxyEdgeAccessResponse, error) {
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	timeout := proxyExitIPTimeout(settings)
	client, err := a.runtime.checkProxyHTTPClient(ctx, req.GetListenerId(), timeout)
	if err != nil {
		return nil, err
	}
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		probeCtx, cancel := context.WithTimeout(ctx, timeout)
		exitIP, err := a.runtime.probeExitIP(probeCtx, client)
		cancel()
		if err != nil {
			return nil, err
		}
		ip = exitIP
	}
	outcome := a.runtime.runEdgeCanary(ctx, client, settings)
	check := buildEdgeAccessCheck(edgeBaseFraudCheck(ip), strings.TrimSpace(req.GetExpectedCountryCode()), outcome)
	return &proxyruntimev1.CheckProxyEdgeAccessResponse{Check: check}, nil
}

func (a runtimeCheckApplication) CheckProxyTargetConnectivity(ctx context.Context, req *proxyruntimev1.CheckProxyTargetConnectivityRequest) (*proxyruntimev1.CheckProxyTargetConnectivityResponse, error) {
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	client, err := a.runtime.checkProxyHTTPClient(ctx, req.GetListenerId(), proxyExitIPTimeout(settings))
	if err != nil {
		return nil, err
	}
	target, err := normalizeConnectivityTarget(req.GetTargetUrl())
	if err != nil {
		return nil, err
	}
	started := time.Now()
	checkReq, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	checkReq.Header.Set("Accept", "text/html,application/json,text/plain;q=0.8")
	checkReq.Header.Set("Cache-Control", "no-cache")
	resp, err := client.Do(checkReq)
	latency := uint32(time.Since(started).Milliseconds())
	check := &proxyruntimev1.ProxyTargetConnectivityCheck{TargetUrl: target, Host: checkReq.URL.Hostname(), LatencyMs: latency, CheckedAt: timestamppb.Now()}
	if err != nil {
		check.ErrorMessage = "target connectivity check failed"
		return &proxyruntimev1.CheckProxyTargetConnectivityResponse{Check: check}, nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	check.StatusCode = uint32(resp.StatusCode)
	check.Reachable = true
	return &proxyruntimev1.CheckProxyTargetConnectivityResponse{Check: check}, nil
}
