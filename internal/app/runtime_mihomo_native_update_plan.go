package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

type mihomoNativeUpdatePlan struct {
	Config               mihomoNativeConfigFile
	ResourceReplacements map[string]mihomoNativeResourceReplacement
}

func buildMihomoNativeUpdatePlan(current mihomoNativeConfigFile, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (mihomoNativeUpdatePlan, error) {
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	plan := mihomoNativeUpdatePlan{
		Config: mihomoNativeConfigFile{
			FixedProxies:   make([]mihomoNativeFixedProxy, 0, len(view.FixedProxies)),
			Proxies:        make([]map[string]any, 0, len(view.FixedProxies)),
			ProxyProviders: map[string]mihomoNativeProvider{},
			ProxyGroups:    preserveNativeGroups(current.ProxyGroups),
			Rules:          append([]string(nil), current.Rules...),
		},
		ResourceReplacements: map[string]mihomoNativeResourceReplacement{},
	}
	if err := plan.addFixedProxies(current, view.GetFixedProxies()); err != nil {
		return mihomoNativeUpdatePlan{}, err
	}
	if err := plan.addSubscriptions(current, view.GetSubscriptions()); err != nil {
		return mihomoNativeUpdatePlan{}, err
	}
	return plan, nil
}
