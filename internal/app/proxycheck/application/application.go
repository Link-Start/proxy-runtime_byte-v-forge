package application

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	"github.com/byte-v-forge/proxy-gateway/internal/app/proxycheck"
)

type Service struct {
	loadSettingsFn func(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
	checkClient    HTTPClientFactory
	probeExitIP    ExitIPProbe
	lookupGeo      GeoLookup
	checkFraud     FraudChecker
	runEdgeCanary  EdgeCanary
	exitCheckCache Cache
	exitIPSF       *singleflight.Group
}

type Dependencies struct {
	LoadSettings   func(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
	CheckClient    HTTPClientFactory
	ProbeExitIP    ExitIPProbe
	LookupGeo      GeoLookup
	CheckFraud     FraudChecker
	RunEdgeCanary  EdgeCanary
	ExitCheckCache Cache
}

type HTTPClientFactory func(context.Context, string, time.Duration) (*http.Client, error)

type ExitIPProbe func(context.Context, *http.Client) (string, error)

type GeoLookup func(context.Context, string) (proxycheck.ExitGeo, error)

type FraudChecker func(context.Context, string, *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyIPFraudCheck, error)

type EdgeCanary func(context.Context, *http.Client, *proxygatewayv1.ProxyGatewayPersistentSettings) proxycheck.EdgeCanaryOutcome

type Cache interface {
	PutExitIP(string, *proxygatewayv1.ProxyExitIP)
	PutGeo(*proxygatewayv1.ProxyExitGeo)
	PutFraud(*proxygatewayv1.ProxyIPFraudCheck)
	PutEdge(string, *proxygatewayv1.ProxyEdgeAccessCheck)
	Snapshot(string) *proxygatewayv1.ProxyExitCheckSnapshot
}

func New(deps Dependencies) Service {
	return Service{
		loadSettingsFn: deps.LoadSettings,
		checkClient:    deps.CheckClient,
		probeExitIP:    deps.ProbeExitIP,
		lookupGeo:      deps.LookupGeo,
		checkFraud:     deps.CheckFraud,
		runEdgeCanary:  deps.RunEdgeCanary,
		exitCheckCache: deps.ExitCheckCache,
		exitIPSF:       &singleflight.Group{},
	}
}

func (a Service) GetProxyExitIP(ctx context.Context, req *proxygatewayv1.GetProxyExitIPRequest) (*proxygatewayv1.GetProxyExitIPResponse, error) {
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	timeout := kernel.ProxyExitIPTimeout(settings)
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
	exitIP := &proxygatewayv1.ProxyExitIP{Ip: ip, CheckedAt: timestamppb.Now()}
	a.putExitIP(req.GetListenerId(), exitIP)
	return &proxygatewayv1.GetProxyExitIPResponse{ProxyExitIp: exitIP}, nil
}

func (a Service) GetProxyExitGeo(ctx context.Context, req *proxygatewayv1.GetProxyExitGeoRequest) (*proxygatewayv1.GetProxyExitGeoResponse, error) {
	ip := strings.TrimSpace(req.GetIp())
	if net.ParseIP(ip) == nil {
		return nil, errors.New("ip is required")
	}
	geo, err := a.lookupExitGeo(ctx, ip)
	if err != nil {
		return nil, err
	}
	out := &proxygatewayv1.ProxyExitGeo{Ip: ip, CountryCode: geo.CountryCode, Region: geo.Region, City: geo.City, CheckedAt: timestamppb.Now()}
	a.putGeo(out)
	return &proxygatewayv1.GetProxyExitGeoResponse{ProxyExitGeo: out}, nil
}

func (a Service) CheckProxyIPFraud(ctx context.Context, req *proxygatewayv1.CheckProxyIPFraudRequest) (*proxygatewayv1.CheckProxyIPFraudResponse, error) {
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
	return &proxygatewayv1.CheckProxyIPFraudResponse{Check: check}, nil
}

func (a Service) CheckProxyEdgeAccess(ctx context.Context, req *proxygatewayv1.CheckProxyEdgeAccessRequest) (*proxygatewayv1.CheckProxyEdgeAccessResponse, error) {
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	timeout := kernel.ProxyExitIPTimeout(settings)
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
	check := proxycheck.BuildEdgeAccessCheck(proxycheck.EdgeBaseFraudCheck(ip), strings.TrimSpace(req.GetExpectedCountryCode()), outcome)
	a.putEdge(req.GetListenerId(), check)
	return &proxygatewayv1.CheckProxyEdgeAccessResponse{Check: check}, nil
}

func (a Service) GetProxyExitCheckSnapshot(_ context.Context, req *proxygatewayv1.GetProxyExitCheckSnapshotRequest) (*proxygatewayv1.GetProxyExitCheckSnapshotResponse, error) {
	return &proxygatewayv1.GetProxyExitCheckSnapshotResponse{Snapshot: a.snapshot(req.GetListenerId())}, nil
}

func (a Service) CheckProxyTargetConnectivity(ctx context.Context, req *proxygatewayv1.CheckProxyTargetConnectivityRequest) (*proxygatewayv1.CheckProxyTargetConnectivityResponse, error) {
	settings, err := a.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	client, err := a.newCheckClient(ctx, req.GetListenerId(), kernel.ProxyExitIPTimeout(settings))
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	target, err := proxycheck.NormalizeConnectivityTarget(req.GetTargetUrl())
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
	check := &proxygatewayv1.ProxyTargetConnectivityCheck{TargetUrl: target, Host: checkReq.URL.Hostname(), LatencyMs: latency, CheckedAt: timestamppb.Now()}
	if err != nil {
		check.ErrorMessage = "target connectivity check failed"
		return &proxygatewayv1.CheckProxyTargetConnectivityResponse{Check: check}, nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	check.StatusCode = uint32(resp.StatusCode)
	check.Reachable = true
	return &proxygatewayv1.CheckProxyTargetConnectivityResponse{Check: check}, nil
}

func (a Service) loadSettings(ctx context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
	if a.loadSettingsFn == nil {
		return nil, appcore.InternalError("runtime check settings repository is not configured", nil)
	}
	return a.loadSettingsFn(ctx)
}

func (a Service) newCheckClient(ctx context.Context, listenerID string, timeout time.Duration) (*http.Client, error) {
	if a.checkClient == nil {
		return nil, appcore.InternalError("runtime check HTTP client factory is not configured", nil)
	}
	return a.checkClient(ctx, listenerID, timeout)
}

func (a Service) checkExitIP(ctx context.Context, client *http.Client) (string, error) {
	if a.probeExitIP == nil {
		return "", appcore.InternalError("runtime check exit IP probe is not configured", nil)
	}
	return a.probeExitIP(ctx, client)
}

func (a Service) checkExitIPDedup(ctx context.Context, listenerID string, client *http.Client) (string, error) {
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

func (a Service) lookupExitGeo(ctx context.Context, ip string) (proxycheck.ExitGeo, error) {
	if a.lookupGeo == nil {
		return proxycheck.ExitGeo{}, appcore.InternalError("runtime check geo lookup is not configured", nil)
	}
	return a.lookupGeo(ctx, ip)
}

func (a Service) checkIPFraud(ctx context.Context, ip string, settings *proxygatewayv1.ProxyGatewayPersistentSettings) (*proxygatewayv1.ProxyIPFraudCheck, error) {
	if a.checkFraud == nil {
		return nil, appcore.InternalError("runtime check IP fraud service is not configured", nil)
	}
	return a.checkFraud(ctx, ip, settings)
}

func (a Service) checkEdgeAccess(ctx context.Context, client *http.Client, settings *proxygatewayv1.ProxyGatewayPersistentSettings) (proxycheck.EdgeCanaryOutcome, error) {
	if a.runEdgeCanary == nil {
		return proxycheck.EdgeCanaryOutcome{}, appcore.InternalError("runtime check edge canary is not configured", nil)
	}
	return a.runEdgeCanary(ctx, client, settings), nil
}

func (a Service) putExitIP(listenerID string, exitIP *proxygatewayv1.ProxyExitIP) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutExitIP(listenerID, exitIP)
	}
}

func (a Service) putGeo(geo *proxygatewayv1.ProxyExitGeo) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutGeo(geo)
	}
}

func (a Service) putFraud(check *proxygatewayv1.ProxyIPFraudCheck) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutFraud(check)
	}
}

func (a Service) putEdge(listenerID string, check *proxygatewayv1.ProxyEdgeAccessCheck) {
	if a.exitCheckCache != nil {
		a.exitCheckCache.PutEdge(listenerID, check)
	}
}

func (a Service) snapshot(listenerID string) *proxygatewayv1.ProxyExitCheckSnapshot {
	if a.exitCheckCache == nil {
		return nil
	}
	return a.exitCheckCache.Snapshot(listenerID)
}
