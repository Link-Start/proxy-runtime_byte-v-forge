package app

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

type runtimeCheckApplication struct {
	settings       runtimeCheckSettings
	checkClient    runtimeCheckHTTPClientFactory
	probeExitIP    runtimeCheckExitIPProbe
	lookupGeo      runtimeCheckGeoLookup
	checkFraud     runtimeCheckFraudChecker
	runEdgeCanary  runtimeCheckEdgeCanary
	exitCheckCache runtimeCheckCache
	exitIPSF       *singleflight.Group
}

type runtimeCheckApplicationDependencies struct {
	Settings       runtimeCheckSettings
	CheckClient    runtimeCheckHTTPClientFactory
	ProbeExitIP    runtimeCheckExitIPProbe
	LookupGeo      runtimeCheckGeoLookup
	CheckFraud     runtimeCheckFraudChecker
	RunEdgeCanary  runtimeCheckEdgeCanary
	ExitCheckCache runtimeCheckCache
}

type runtimeCheckSettings interface {
	load(context.Context) (*runtimeSettingsFile, error)
}

type runtimeCheckHTTPClientFactory func(context.Context, string, time.Duration) (*http.Client, error)

type runtimeCheckExitIPProbe func(context.Context, *http.Client) (string, error)

type runtimeCheckGeoLookup func(context.Context, string) (proxycheck.ExitGeo, error)

type runtimeCheckFraudChecker func(context.Context, string, *runtimeSettingsFile) (*proxyruntimev1.ProxyIPFraudCheck, error)

type runtimeCheckEdgeCanary func(context.Context, *http.Client, *runtimeSettingsFile) edgeCanaryOutcome

type runtimeCheckCache interface {
	PutExitIP(string, *proxyruntimev1.ProxyExitIP)
	PutGeo(*proxyruntimev1.ProxyExitGeo)
	PutFraud(*proxyruntimev1.ProxyIPFraudCheck)
	PutEdge(string, *proxyruntimev1.ProxyEdgeAccessCheck)
	Snapshot(string) *proxyruntimev1.ProxyExitCheckSnapshot
}

func newRuntimeCheckApplication(deps runtimeCheckApplicationDependencies) runtimeCheckApplication {
	return runtimeCheckApplication{
		settings:       deps.Settings,
		checkClient:    deps.CheckClient,
		probeExitIP:    deps.ProbeExitIP,
		lookupGeo:      deps.LookupGeo,
		checkFraud:     deps.CheckFraud,
		runEdgeCanary:  deps.RunEdgeCanary,
		exitCheckCache: deps.ExitCheckCache,
		exitIPSF:       &singleflight.Group{},
	}
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

func (s *RuntimeService) GetProxyExitCheckSnapshot(ctx context.Context, req *proxyruntimev1.GetProxyExitCheckSnapshotRequest) (*proxyruntimev1.GetProxyExitCheckSnapshotResponse, error) {
	return s.checks.GetProxyExitCheckSnapshot(ctx, req)
}

func (a runtimeCheckApplication) GetProxyExitIP(ctx context.Context, req *proxyruntimev1.GetProxyExitIPRequest) (*proxyruntimev1.GetProxyExitIPResponse, error) {
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	timeout := settingscore.ProxyExitIPTimeout(settings)
	client, err := a.newCheckClient(ctx, req.GetListenerId(), timeout)
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ip, err := a.checkExitIPDedup(probeCtx, req.GetListenerId(), client)
	if err != nil {
		return nil, err
	}
	exitIP := &proxyruntimev1.ProxyExitIP{Ip: ip, CheckedAt: timestamppb.Now()}
	a.putExitIP(req.GetListenerId(), exitIP)
	return &proxyruntimev1.GetProxyExitIPResponse{ProxyExitIp: exitIP}, nil
}

func (a runtimeCheckApplication) GetProxyExitGeo(ctx context.Context, req *proxyruntimev1.GetProxyExitGeoRequest) (*proxyruntimev1.GetProxyExitGeoResponse, error) {
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		return nil, errors.New("ip is required")
	}
	geo, err := a.lookupExitGeo(ctx, ip)
	if err != nil {
		return nil, err
	}
	out := &proxyruntimev1.ProxyExitGeo{Ip: ip, CountryCode: geo.CountryCode, Region: geo.Region, City: geo.City, CheckedAt: timestamppb.Now()}
	a.putGeo(out)
	return &proxyruntimev1.GetProxyExitGeoResponse{ProxyExitGeo: out}, nil
}

func (a runtimeCheckApplication) CheckProxyIPFraud(ctx context.Context, req *proxyruntimev1.CheckProxyIPFraudRequest) (*proxyruntimev1.CheckProxyIPFraudResponse, error) {
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		return nil, errors.New("ip is required")
	}
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	check, err := a.checkIPFraud(ctx, ip, settings)
	if err != nil {
		if appcore.IsAppError(err) {
			return nil, err
		}
		return nil, errors.New("check IP fraud")
	}
	a.putFraud(check)
	return &proxyruntimev1.CheckProxyIPFraudResponse{Check: check}, nil
}

func (a runtimeCheckApplication) CheckProxyEdgeAccess(ctx context.Context, req *proxyruntimev1.CheckProxyEdgeAccessRequest) (*proxyruntimev1.CheckProxyEdgeAccessResponse, error) {
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	timeout := settingscore.ProxyExitIPTimeout(settings)
	client, err := a.newCheckClient(ctx, req.GetListenerId(), timeout)
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		probeCtx, cancel := context.WithTimeout(ctx, timeout)
		exitIP, err := a.checkExitIPDedup(probeCtx, req.GetListenerId(), client)
		cancel()
		if err != nil {
			return nil, err
		}
		ip = exitIP
	}
	outcome, err := a.checkEdgeAccess(ctx, client, settings)
	if err != nil {
		return nil, err
	}
	check := buildEdgeAccessCheck(edgeBaseFraudCheck(ip), strings.TrimSpace(req.GetExpectedCountryCode()), outcome)
	a.putEdge(req.GetListenerId(), check)
	return &proxyruntimev1.CheckProxyEdgeAccessResponse{Check: check}, nil
}

func (a runtimeCheckApplication) GetProxyExitCheckSnapshot(_ context.Context, req *proxyruntimev1.GetProxyExitCheckSnapshotRequest) (*proxyruntimev1.GetProxyExitCheckSnapshotResponse, error) {
	return &proxyruntimev1.GetProxyExitCheckSnapshotResponse{Snapshot: a.snapshot(req.GetListenerId())}, nil
}

func (a runtimeCheckApplication) CheckProxyTargetConnectivity(ctx context.Context, req *proxyruntimev1.CheckProxyTargetConnectivityRequest) (*proxyruntimev1.CheckProxyTargetConnectivityResponse, error) {
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	client, err := a.newCheckClient(ctx, req.GetListenerId(), settingscore.ProxyExitIPTimeout(settings))
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	target, err := normalizeConnectivityTarget(req.GetTargetUrl())
	if err != nil {
		return nil, err
	}
	started := time.Now()
	checkReq, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	checkReq.Close = true
	checkReq.Header.Set("Accept", "text/html,application/json,text/plain;q=0.8")
	checkReq.Header.Set("Cache-Control", "no-cache")
	checkReq.Header.Set("Connection", "close")
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

func (a runtimeCheckApplication) loadSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	if a.settings == nil {
		return nil, appcore.InternalError("runtime check settings repository is not configured", nil)
	}
	return a.settings.load(ctx)
}

func (a runtimeCheckApplication) newCheckClient(ctx context.Context, listenerID string, timeout time.Duration) (*http.Client, error) {
	if a.checkClient == nil {
		return nil, appcore.InternalError("runtime check HTTP client factory is not configured", nil)
	}
	return a.checkClient(ctx, listenerID, timeout)
}

func (a runtimeCheckApplication) checkExitIP(ctx context.Context, client *http.Client) (string, error) {
	if a.probeExitIP == nil {
		return "", appcore.InternalError("runtime check exit IP probe is not configured", nil)
	}
	return a.probeExitIP(ctx, client)
}

func (a runtimeCheckApplication) checkExitIPDedup(ctx context.Context, listenerID string, client *http.Client) (string, error) {
	if a.exitIPSF == nil {
		return a.checkExitIP(ctx, client)
	}
	ip, err, _ := a.exitIPSF.Do(listenerID, func() (any, error) {
		return a.checkExitIP(ctx, client)
	})
	if err != nil {
		return "", err
	}
	return ip.(string), nil
}

func (a runtimeCheckApplication) lookupExitGeo(ctx context.Context, ip string) (proxycheck.ExitGeo, error) {
	if a.lookupGeo == nil {
		return proxycheck.ExitGeo{}, appcore.InternalError("runtime check geo lookup is not configured", nil)
	}
	return a.lookupGeo(ctx, ip)
}

func (a runtimeCheckApplication) checkIPFraud(ctx context.Context, ip string, settings *runtimeSettingsFile) (*proxyruntimev1.ProxyIPFraudCheck, error) {
	if a.checkFraud == nil {
		return nil, appcore.InternalError("runtime check IP fraud service is not configured", nil)
	}
	return a.checkFraud(ctx, ip, settings)
}

func (a runtimeCheckApplication) checkEdgeAccess(ctx context.Context, client *http.Client, settings *runtimeSettingsFile) (edgeCanaryOutcome, error) {
	if a.runEdgeCanary == nil {
		return edgeCanaryOutcome{}, appcore.InternalError("runtime check edge canary is not configured", nil)
	}
	return a.runEdgeCanary(ctx, client, settings), nil
}

func (a runtimeCheckApplication) putExitIP(listenerID string, exitIP *proxyruntimev1.ProxyExitIP) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutExitIP(listenerID, exitIP)
	}
}

func (a runtimeCheckApplication) putGeo(geo *proxyruntimev1.ProxyExitGeo) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutGeo(geo)
	}
}

func (a runtimeCheckApplication) putFraud(check *proxyruntimev1.ProxyIPFraudCheck) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutFraud(check)
	}
}

func (a runtimeCheckApplication) putEdge(listenerID string, check *proxyruntimev1.ProxyEdgeAccessCheck) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutEdge(listenerID, check)
	}
}

func (a runtimeCheckApplication) snapshot(listenerID string) *proxyruntimev1.ProxyExitCheckSnapshot {
	if a.exitCheckCache == nil {
		return nil
	}
	return a.exitCheckCache.Snapshot(listenerID)
}
