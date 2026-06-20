package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	settingssecret "github.com/byte-v-forge/proxy-runtime/internal/app/settings/adapter/secret"
)

func (r *Runtime) lookupIPGeo(ctx context.Context, ip string) (proxycheck.ExitGeo, error) {
	if geo, ok := r.geoCache.Get(ip); ok {
		return geo, nil
	}
	out, err, _ := r.geoLookupSF.Do(ip, func() (any, error) {
		if geo, ok := r.geoCache.Get(ip); ok {
			return geo, nil
		}
		return r.loadIPGeo(ctx, ip)
	})
	if err != nil {
		return proxycheck.ExitGeo{}, err
	}
	return out.(proxycheck.ExitGeo), nil
}

func (r *Runtime) loadIPGeo(ctx context.Context, ip string) (proxycheck.ExitGeo, error) {
	settings, err := r.settings.load(ctx)
	if err != nil {
		return proxycheck.ExitGeo{}, err
	}
	providers, err := settingssecret.IPGeoProviders(ctx, r.store, settings, r.ipGeoProviders)
	if err != nil {
		return proxycheck.ExitGeo{}, err
	}
	if len(providers) == 0 {
		return proxycheck.ExitGeo{}, errors.New("IP geo provider is not configured")
	}
	lookupCtx, cancel := context.WithTimeout(ctx, kernel.ProxyExitIPTimeout(settings))
	defer cancel()
	geo, err := newIPGeoLookup(r.ipGeoProviders, kernel.ProxyExitIPTimeout(settings), providers, r.logger).Lookup(lookupCtx, ip)
	if err != nil {
		return proxycheck.ExitGeo{}, err
	}
	out := proxycheck.ExitGeoFromProto(ip, geo)
	r.geoCache.Put(ip, out)
	return out, nil
}

func newIPGeoLookup(registry *ipgeo.Registry, timeout time.Duration, providers []ipgeo.ProviderConfig, logger *slog.Logger) *ipgeo.Service {
	return ipgeo.NewService(registry, ipgeo.Config{Providers: providers, Timeout: timeout}, logger)
}
