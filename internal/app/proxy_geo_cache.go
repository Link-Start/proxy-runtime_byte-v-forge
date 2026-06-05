package app

import (
	"strings"
	"sync"
	"time"
)

type ipGeoCache struct {
	mu    sync.Mutex
	items map[string]cachedIPGeo
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
	if !ok || time.Now().After(item.expiresAt) {
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
	c.items[ip] = cachedIPGeo{geo: geo, expiresAt: time.Now().Add(ipGeoCacheTTL)}
}
