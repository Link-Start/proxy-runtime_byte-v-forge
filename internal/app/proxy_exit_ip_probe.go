package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
)

func (r *Runtime) probeExitIP(ctx context.Context, client *http.Client) (string, error) {
	endpoints := append([]string(nil), r.cfg.ProxyExitGeoURLs...)
	if len(endpoints) == 0 {
		return "", errors.New("proxy exit ip endpoints are not configured")
	}
	probeCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	type probeResult struct {
		ip  string
		err error
	}
	results := make(chan probeResult, len(endpoints))
	for _, endpoint := range endpoints {
		endpoint := strings.TrimSpace(endpoint)
		if endpoint == "" {
			results <- probeResult{err: errors.New("empty proxy exit ip endpoint")}
			continue
		}
		go func() {
			geo, err := proxycheck.RequestIPInfo(probeCtx, client, endpoint, true)
			if err != nil {
				results <- probeResult{err: err}
				return
			}
			if net.ParseIP(geo.IP) == nil {
				results <- probeResult{err: errors.New("proxy exit ip endpoint returned invalid IP")}
				return
			}
			results <- probeResult{ip: geo.IP}
		}()
	}
	var probeErrs []error
	for range endpoints {
		select {
		case <-ctx.Done():
			return "", errors.New("check proxy exit ip timed out")
		case result := <-results:
			if result.err == nil && result.ip != "" {
				cancel()
				return result.ip, nil
			}
			if result.err != nil {
				probeErrs = append(probeErrs, result.err)
			}
		}
	}
	if len(probeErrs) > 0 {
		return "", fmt.Errorf("check proxy exit ip failed: %w", errors.Join(probeErrs...))
	}
	return "", errors.New("check proxy exit ip failed")
}
