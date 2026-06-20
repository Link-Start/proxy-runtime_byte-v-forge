package dynamic

import (
	"context"
	"net"
	"net/url"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/geox"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (p *IPSelector) endpointRegionCodes(ctx context.Context, endpoint accountproxy.Gateway, policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy) []string {
	if !hasRequestedRegion(policy) {
		return nil
	}
	host := networkAddressHost(endpoint.EndpointURL)
	out := regionCodesFromEndpointHost(host)
	if host == "" || p.lookupIPGeo == nil {
		return cleanRegionCodes(out)
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if ip := net.ParseIP(host); ip != nil {
		return p.endpointRegionCodesFromIP(lookupCtx, out, ip)
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err != nil {
		return cleanRegionCodes(out)
	}
	for _, addr := range addrs {
		if addr.IP == nil || !publicIP(addr.IP) {
			continue
		}
		return p.endpointRegionCodesFromIP(lookupCtx, out, addr.IP)
	}
	return cleanRegionCodes(out)
}

func (p *IPSelector) endpointRegionCodesFromIP(ctx context.Context, base []string, ip net.IP) []string {
	if ip == nil || !publicIP(ip) {
		return cleanRegionCodes(base)
	}
	geo, err := p.lookupIPGeo(ctx, ip.String())
	if err != nil {
		return cleanRegionCodes(base)
	}
	out := appendRegionCode(base, geox.NormalizeCountryAlpha2(geo.CountryCode))
	out = appendRegionCode(out, geox.NormalizeRegionCode(geo.Region))
	if region := geox.CountryRegionCode(geo.CountryCode); region != "" {
		out = appendRegionCode(out, region)
	}
	return cleanRegionCodes(out)
}

func regionCodesFromEndpointHost(host string) []string {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil
	}
	out := []string{}
	for _, country := range geox.CountryCodesInText(host) {
		out = appendRegionCode(out, country)
		if region := geox.CountryRegionCode(country); region != "" {
			out = appendRegionCode(out, region)
		}
	}
	for _, token := range strings.FieldsFunc(host, func(r rune) bool {
		return !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	}) {
		out = appendRegionCode(out, geox.NormalizeRegionCode(token))
	}
	return cleanRegionCodes(out)
}

func publicIP(ip net.IP) bool {
	return ip != nil && !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified()
}

func networkAddressHost(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return strings.Trim(host, "[]")
	}
	if parsed, err := url.Parse(addr); err == nil && parsed != nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return strings.Trim(addr, "[]")
}
