package app

import (
	"strings"
	"sync"

	"github.com/byte-v-forge/proxy-runtime/internal/clock"
)

type ipGeoCache struct {
	mu    sync.Mutex
	items map[string]cachedIPGeo
	clock clock.Clock
}

func (c *ipGeoCache) get(ip string) (proxyExitGeo, bool) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return proxyExitGeo{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		return proxyExitGeo{}, false
	}
	item, ok := c.items[ip]
	if !ok || c.clock.Now().After(item.expiresAt) {
		delete(c.items, ip)
		return proxyExitGeo{}, false
	}
	return item.geo, true
}

func (c *ipGeoCache) put(ip string, geo proxyExitGeo) {
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

func (c *ipGeoCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = nil
}
