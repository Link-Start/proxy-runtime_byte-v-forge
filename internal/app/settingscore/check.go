package settingscore

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
