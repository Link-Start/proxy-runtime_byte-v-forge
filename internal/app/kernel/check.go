package kernel

import (
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const DefaultProxyExitIPTimeout = 5 * time.Second

func NormalizeCheckSettings(settings *proxygatewayv1.ProxyGatewayCheckSettings) *proxygatewayv1.ProxyGatewayCheckSettings {
	if settings == nil {
		settings = &proxygatewayv1.ProxyGatewayCheckSettings{}
	}
	if settings.GetProxyExitIpTimeout().AsDuration() <= 0 {
		settings.ProxyExitIpTimeout = durationpb.New(DefaultProxyExitIPTimeout)
	}
	return settings
}

func ProxyExitIPTimeout(settings *proxygatewayv1.ProxyGatewayPersistentSettings) time.Duration {
	if settings == nil {
		return DefaultProxyExitIPTimeout
	}
	duration := NormalizeCheckSettings(settings.GetCheckSettings()).GetProxyExitIpTimeout().AsDuration()
	if duration <= 0 {
		return DefaultProxyExitIPTimeout
	}
	return duration
}

func CheckSettingsFromRequest(req *proxygatewayv1.ProxyGatewayCheckSettings, current *proxygatewayv1.ProxyGatewayCheckSettings) *proxygatewayv1.ProxyGatewayCheckSettings {
	if req == nil {
		return CloneCheckSettings(current)
	}
	return NormalizeCheckSettings(&proxygatewayv1.ProxyGatewayCheckSettings{ProxyExitIpTimeout: req.GetProxyExitIpTimeout()})
}

func CloneCheckSettings(in *proxygatewayv1.ProxyGatewayCheckSettings) *proxygatewayv1.ProxyGatewayCheckSettings {
	in = NormalizeCheckSettings(in)
	return &proxygatewayv1.ProxyGatewayCheckSettings{ProxyExitIpTimeout: durationpb.New(in.GetProxyExitIpTimeout().AsDuration())}
}
