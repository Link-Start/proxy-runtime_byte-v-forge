package mihomonative

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func SettingsEmpty(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) bool {
	return view == nil || len(view.GetFixedProxies()) == 0 && len(view.GetSubscriptions()) == 0
}
