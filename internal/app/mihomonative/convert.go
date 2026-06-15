package mihomonative

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func SettingsFromConfig(config ConfigFile) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	view := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{
		FixedProxies:  FixedProxySettingsFromConfig(config),
		Subscriptions: SubscriptionSettingsFromConfig(config),
	}
	return NormalizeSettings(view)
}
