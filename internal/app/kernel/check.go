package kernel

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const DefaultProxyExitIPTimeout = 5 * time.Second

func NormalizeCheckSettings(settings *proxyruntimev1.ProxyRuntimeCheckSettings) *proxyruntimev1.ProxyRuntimeCheckSettings {
	if settings == nil {
		settings = &proxyruntimev1.ProxyRuntimeCheckSettings{}
	}
	if settings.GetProxyExitIpTimeout().AsDuration() <= 0 {
		settings.ProxyExitIpTimeout = durationpb.New(DefaultProxyExitIPTimeout)
	}
	return settings
}

func ProxyExitIPTimeout(settings *proxyruntimev1.ProxyRuntimePersistentSettings) time.Duration {
	if settings == nil {
		return DefaultProxyExitIPTimeout
	}
	duration := NormalizeCheckSettings(settings.GetCheckSettings()).GetProxyExitIpTimeout().AsDuration()
	if duration <= 0 {
		return DefaultProxyExitIPTimeout
	}
	return duration
}

func CheckSettingsFromRequest(req *proxyruntimev1.ProxyRuntimeCheckSettings, current *proxyruntimev1.ProxyRuntimeCheckSettings) *proxyruntimev1.ProxyRuntimeCheckSettings {
	if req == nil {
		return CloneCheckSettings(current)
	}
	return NormalizeCheckSettings(&proxyruntimev1.ProxyRuntimeCheckSettings{ProxyExitIpTimeout: req.GetProxyExitIpTimeout()})
}

func CloneCheckSettings(in *proxyruntimev1.ProxyRuntimeCheckSettings) *proxyruntimev1.ProxyRuntimeCheckSettings {
	in = NormalizeCheckSettings(in)
	return &proxyruntimev1.ProxyRuntimeCheckSettings{ProxyExitIpTimeout: durationpb.New(in.GetProxyExitIpTimeout().AsDuration())}
}
