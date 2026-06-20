package proxycheck

import (
	"strings"
	"sync"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/clock"
)

const ipGeoCacheTTL = 24 * time.Hour

type cachedIPGeo struct {
	geo       ExitGeo
	expiresAt time.Time
}

type IPGeoCache struct {
	mu    sync.Mutex
	items map[string]cachedIPGeo
	clock clock.Clock
}

func NewIPGeoCache(clk clock.Clock) *IPGeoCache {
	return &IPGeoCache{clock: clk}
}

func (c *IPGeoCache) Get(ip string) (ExitGeo, bool) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ExitGeo{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		return ExitGeo{}, false
	}
	item, ok := c.items[ip]
	if !ok || c.clock.Now().After(item.expiresAt) {
		delete(c.items, ip)
		return ExitGeo{}, false
	}
	return item.geo, true
}

func (c *IPGeoCache) Put(ip string, geo ExitGeo) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		c.items = map[string]cachedIPGeo{}
	}
	c.items[ip] = cachedIPGeo{geo: geo, expiresAt: c.clock.Now().Add(ipGeoCacheTTL)}
}

func (c *IPGeoCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = nil
}
