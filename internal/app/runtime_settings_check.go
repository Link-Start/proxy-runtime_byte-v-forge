package app

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func proxyExitIPTimeout(settings *runtimeSettingsFile) time.Duration {
	if settings == nil {
		return settingscore.DefaultProxyExitIPTimeout
	}
	duration := settingscore.NormalizeCheckSettings(settings.GetCheckSettings()).GetProxyExitIpTimeout().AsDuration()
	if duration <= 0 {
		return settingscore.DefaultProxyExitIPTimeout
	}
	return duration
}

func checkSettingsFromRequest(req *proxyruntimev1.ProxyRuntimeCheckSettings, current *proxyruntimev1.ProxyRuntimeCheckSettings) *proxyruntimev1.ProxyRuntimeCheckSettings {
	if req == nil {
		return cloneCheckSettings(current)
	}
	return settingscore.NormalizeCheckSettings(&proxyruntimev1.ProxyRuntimeCheckSettings{ProxyExitIpTimeout: req.GetProxyExitIpTimeout()})
}

func cloneCheckSettings(in *proxyruntimev1.ProxyRuntimeCheckSettings) *proxyruntimev1.ProxyRuntimeCheckSettings {
	in = settingscore.NormalizeCheckSettings(in)
	return &proxyruntimev1.ProxyRuntimeCheckSettings{ProxyExitIpTimeout: durationpb.New(in.GetProxyExitIpTimeout().AsDuration())}
}
