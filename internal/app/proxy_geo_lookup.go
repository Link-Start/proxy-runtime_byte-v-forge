package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
)

func (r *Runtime) lookupIPGeo(ctx context.Context, ip string) (proxyExitGeo, error) {
	if geo, ok := r.geoCache.get(ip); ok {
		return geo, nil
	}
	out, err, _ := r.geoLookupSF.Do(ip, func() (any, error) {
		if geo, ok := r.geoCache.get(ip); ok {
			return geo, nil
		}
		return r.loadIPGeo(ctx, ip)
	})
	if err != nil {
		return proxyExitGeo{}, err
	}
	return out.(proxyExitGeo), nil
}

func (r *Runtime) loadIPGeo(ctx context.Context, ip string) (proxyExitGeo, error) {
	settings, err := r.settings.load(ctx)
	if err != nil {
		return proxyExitGeo{}, err
	}
	providers, err := ipGeoProviders(ctx, r.store, settings, r.ipGeoProviders)
	if err != nil {
		return proxyExitGeo{}, err
	}
	if len(providers) == 0 {
		return proxyExitGeo{}, errors.New("IP geo provider is not configured")
	}
	lookupCtx, cancel := context.WithTimeout(ctx, proxyExitIPTimeout(settings))
	defer cancel()
	geo, err := newIPGeoLookup(r.ipGeoProviders, proxyExitIPTimeout(settings), providers, r.logger).Lookup(lookupCtx, ip)
	if err != nil {
		return proxyExitGeo{}, err
	}
	out := proxyExitGeoFromProto(ip, geo)
	r.geoCache.put(ip, out)
	return out, nil
}

func newIPGeoLookup(registry *ipgeo.Registry, timeout time.Duration, providers []ipgeo.ProviderConfig, logger *slog.Logger) *ipgeo.Service {
	return ipgeo.NewService(registry, ipgeo.Config{Providers: providers, Timeout: timeout}, logger)
}

func proxyExitGeoFromProto(ip string, geo *proxyruntimev1.ProxyExitGeo) proxyExitGeo {
	return proxyExitGeo{
		IP:          firstNonEmpty(geo.GetIp(), ip),
		CountryCode: geo.GetCountryCode(),
		Region:      geo.GetRegion(),
		City:        geo.GetCity(),
	}
}
