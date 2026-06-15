package mihomonative

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

type UpdatePlan struct {
	Config               ConfigFile
	ResourceReplacements map[string]ResourceReplacement
}

func BuildUpdatePlan(current ConfigFile, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (UpdatePlan, error) {
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	plan := UpdatePlan{
		Config: ConfigFile{
			FixedProxies:   make([]FixedProxy, 0, len(view.FixedProxies)),
			Proxies:        make([]map[string]any, 0, len(view.FixedProxies)),
			ProxyProviders: map[string]Provider{},
			ProxyGroups:    preserveGroups(current.ProxyGroups),
			Rules:          append([]string(nil), current.Rules...),
		},
		ResourceReplacements: map[string]ResourceReplacement{},
	}
	if err := plan.addFixedProxies(current, view.GetFixedProxies()); err != nil {
		return UpdatePlan{}, err
	}
	if err := plan.addSubscriptions(current, view.GetSubscriptions()); err != nil {
		return UpdatePlan{}, err
	}
	return plan, nil
}
