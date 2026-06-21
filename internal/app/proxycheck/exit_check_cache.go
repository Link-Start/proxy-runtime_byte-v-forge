package proxycheck

import (
	"strings"
	"sync"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const proxyExitCheckCacheTTL = 24 * time.Hour

type ExitCheckCache struct {
	mu        sync.Mutex
	exitIPs   map[string]cachedProxyExitIP
	geos      map[string]cachedProxyExitGeo
	frauds    map[string]cachedProxyIPFraudCheck
	edgeRisks map[string]cachedProxyEdgeAccessCheck
	clock     clock.Clock
}

func NewExitCheckCache(clk clock.Clock) *ExitCheckCache {
	return &ExitCheckCache{clock: clk}
}

type cachedProxyExitIP struct {
	value     *proxygatewayv1.ProxyExitIP
	updatedAt time.Time
	expiresAt time.Time
}

type cachedProxyExitGeo struct {
	value     *proxygatewayv1.ProxyExitGeo
	updatedAt time.Time
	expiresAt time.Time
}

type cachedProxyIPFraudCheck struct {
	value     *proxygatewayv1.ProxyIPFraudCheck
	updatedAt time.Time
	expiresAt time.Time
}

type cachedProxyEdgeAccessCheck struct {
	value     *proxygatewayv1.ProxyEdgeAccessCheck
	updatedAt time.Time
	expiresAt time.Time
}

func (c *ExitCheckCache) Snapshot(listenerID string) *proxygatewayv1.ProxyExitCheckSnapshot {
	listenerID = normalizeProxyExitCheckCacheKey(listenerID)
	if listenerID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock.Now()
	exitIP, ok := c.cachedExitIP(listenerID, now)
	if !ok {
		return nil
	}
	snapshot := &proxygatewayv1.ProxyExitCheckSnapshot{
		ListenerId:  listenerID,
		ProxyExitIp: cloneProxyExitIP(exitIP.value),
		UpdatedAt:   timestamppb.New(exitIP.updatedAt),
		ExpiresAt:   timestamppb.New(exitIP.expiresAt),
	}
	ip := strings.TrimSpace(exitIP.value.GetIp())
	if geo, ok := c.cachedGeo(ip, now); ok {
		snapshot.ProxyExitGeo = cloneProxyExitGeo(geo.value)
		snapshot.UpdatedAt = maxTimestamp(snapshot.GetUpdatedAt(), geo.updatedAt)
		snapshot.ExpiresAt = minTimestamp(snapshot.GetExpiresAt(), geo.expiresAt)
	}
	if fraud, ok := c.cachedFraud(ip, now); ok {
		snapshot.IpFraudCheck = cloneProxyIPFraudCheck(fraud.value)
		snapshot.UpdatedAt = maxTimestamp(snapshot.GetUpdatedAt(), fraud.updatedAt)
		snapshot.ExpiresAt = minTimestamp(snapshot.GetExpiresAt(), fraud.expiresAt)
	}
	if edge, ok := c.cachedEdge(listenerID, now); ok {
		snapshot.EdgeAccessCheck = cloneProxyEdgeAccessCheck(edge.value)
		snapshot.UpdatedAt = maxTimestamp(snapshot.GetUpdatedAt(), edge.updatedAt)
		snapshot.ExpiresAt = minTimestamp(snapshot.GetExpiresAt(), edge.expiresAt)
	}
	return snapshot
}

func (c *ExitCheckCache) PutExitIP(listenerID string, value *proxygatewayv1.ProxyExitIP) {
	listenerID = normalizeProxyExitCheckCacheKey(listenerID)
	if listenerID == "" || value == nil || strings.TrimSpace(value.GetIp()) == "" {
		return
	}
	now := c.clock.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.exitIPs == nil {
		c.exitIPs = map[string]cachedProxyExitIP{}
	}
	c.exitIPs[listenerID] = cachedProxyExitIP{value: cloneProxyExitIP(value), updatedAt: now, expiresAt: now.Add(proxyExitCheckCacheTTL)}
}

func (c *ExitCheckCache) PutGeo(value *proxygatewayv1.ProxyExitGeo) {
	if value == nil {
		return
	}
	ip := normalizeProxyExitCheckCacheKey(value.GetIp())
	if ip == "" {
		return
	}
	now := c.clock.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.geos == nil {
		c.geos = map[string]cachedProxyExitGeo{}
	}
	c.geos[ip] = cachedProxyExitGeo{value: cloneProxyExitGeo(value), updatedAt: now, expiresAt: now.Add(proxyExitCheckCacheTTL)}
}

func (c *ExitCheckCache) PutFraud(value *proxygatewayv1.ProxyIPFraudCheck) {
	if value == nil {
		return
	}
	ip := normalizeProxyExitCheckCacheKey(value.GetIp())
	if ip == "" {
		return
	}
	now := c.clock.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.frauds == nil {
		c.frauds = map[string]cachedProxyIPFraudCheck{}
	}
	c.frauds[ip] = cachedProxyIPFraudCheck{value: cloneProxyIPFraudCheck(value), updatedAt: now, expiresAt: now.Add(proxyExitCheckCacheTTL)}
}

func (c *ExitCheckCache) PutEdge(listenerID string, value *proxygatewayv1.ProxyEdgeAccessCheck) {
	listenerID = normalizeProxyExitCheckCacheKey(listenerID)
	if listenerID == "" || value == nil {
		return
	}
	now := c.clock.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.edgeRisks == nil {
		c.edgeRisks = map[string]cachedProxyEdgeAccessCheck{}
	}
	c.edgeRisks[listenerID] = cachedProxyEdgeAccessCheck{value: cloneProxyEdgeAccessCheck(value), updatedAt: now, expiresAt: now.Add(proxyExitCheckCacheTTL)}
}

func (c *ExitCheckCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.exitIPs = nil
	c.geos = nil
	c.frauds = nil
	c.edgeRisks = nil
}

func (c *ExitCheckCache) cachedExitIP(listenerID string, now time.Time) (cachedProxyExitIP, bool) {
	item, ok := c.exitIPs[listenerID]
	if !ok || now.After(item.expiresAt) {
		delete(c.exitIPs, listenerID)
		return cachedProxyExitIP{}, false
	}
	return item, true
}

func (c *ExitCheckCache) cachedGeo(ip string, now time.Time) (cachedProxyExitGeo, bool) {
	item, ok := c.geos[normalizeProxyExitCheckCacheKey(ip)]
	if !ok || now.After(item.expiresAt) {
		delete(c.geos, normalizeProxyExitCheckCacheKey(ip))
		return cachedProxyExitGeo{}, false
	}
	return item, true
}

func (c *ExitCheckCache) cachedFraud(ip string, now time.Time) (cachedProxyIPFraudCheck, bool) {
	item, ok := c.frauds[normalizeProxyExitCheckCacheKey(ip)]
	if !ok || now.After(item.expiresAt) {
		delete(c.frauds, normalizeProxyExitCheckCacheKey(ip))
		return cachedProxyIPFraudCheck{}, false
	}
	return item, true
}

func (c *ExitCheckCache) cachedEdge(listenerID string, now time.Time) (cachedProxyEdgeAccessCheck, bool) {
	item, ok := c.edgeRisks[listenerID]
	if !ok || now.After(item.expiresAt) {
		delete(c.edgeRisks, listenerID)
		return cachedProxyEdgeAccessCheck{}, false
	}
	return item, true
}

func normalizeProxyExitCheckCacheKey(value string) string {
	return strings.TrimSpace(value)
}

func cloneProxyExitIP(value *proxygatewayv1.ProxyExitIP) *proxygatewayv1.ProxyExitIP {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*proxygatewayv1.ProxyExitIP)
}

func cloneProxyExitGeo(value *proxygatewayv1.ProxyExitGeo) *proxygatewayv1.ProxyExitGeo {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*proxygatewayv1.ProxyExitGeo)
}

func cloneProxyIPFraudCheck(value *proxygatewayv1.ProxyIPFraudCheck) *proxygatewayv1.ProxyIPFraudCheck {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*proxygatewayv1.ProxyIPFraudCheck)
}

func cloneProxyEdgeAccessCheck(value *proxygatewayv1.ProxyEdgeAccessCheck) *proxygatewayv1.ProxyEdgeAccessCheck {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*proxygatewayv1.ProxyEdgeAccessCheck)
}

func maxTimestamp(current *timestamppb.Timestamp, candidate time.Time) *timestamppb.Timestamp {
	if current == nil || candidate.After(current.AsTime()) {
		return timestamppb.New(candidate)
	}
	return current
}

func minTimestamp(current *timestamppb.Timestamp, candidate time.Time) *timestamppb.Timestamp {
	if current == nil || candidate.Before(current.AsTime()) {
		return timestamppb.New(candidate)
	}
	return current
}
