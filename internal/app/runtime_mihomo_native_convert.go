package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func mihomoNativeSettingsFromConfig(config mihomoNativeConfigFile) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	view := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{
		FixedProxies:  mihomoNativeFixedProxySettingsFromConfig(config),
		Subscriptions: mihomoNativeSubscriptionSettingsFromConfig(config),
	}
	return normalizeMihomoNativeSettings(view)
}
