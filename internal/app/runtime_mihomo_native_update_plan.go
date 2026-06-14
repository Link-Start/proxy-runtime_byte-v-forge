package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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

func (p *mihomoNativeUpdatePlan) addFixedProxies(current mihomoNativeConfigFile, fixedProxies []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) error {
	currentByID, currentByName := currentFixedProxyIndexes(current)
	seenIDs := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	for _, item := range fixedProxies {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(item), currentByName)
		if existing := currentByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			p.ResourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		}
		p.ResourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		if _, exists := seenIDs[normalized.ID]; exists {
			return invalidArgument(fmt.Sprintf("fixed proxy %q duplicates id %q", normalized.Name, normalized.ID), nil)
		}
		if _, exists := seenNames[normalized.Name]; exists {
			return invalidArgument(fmt.Sprintf("fixed proxy %q duplicates name", normalized.Name), nil)
		}
		seenIDs[normalized.ID] = struct{}{}
		seenNames[normalized.Name] = struct{}{}
		proxy, err := mihomoNativeProxyFromURI(normalized.Name, normalized.URI)
		if err != nil {
			return invalidArgument(err.Error(), nil)
		}
		normalized.Type = jsonStringValue(proxy["type"])
		p.Config.FixedProxies = append(p.Config.FixedProxies, normalized)
		p.Config.Proxies = append(p.Config.Proxies, proxy)
	}
	return nil
}

func (p *mihomoNativeUpdatePlan) addSubscriptions(current mihomoNativeConfigFile, subscriptions []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) error {
	currentByID, currentByName := currentSubscriptionIndexes(current)
	seenIDs := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	for _, item := range subscriptions {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(item), currentByName)
		if existing := currentByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			p.ResourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		}
		p.ResourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		if _, exists := seenIDs[normalized.ID]; exists {
			return invalidArgument(fmt.Sprintf("subscription %q duplicates id %q", normalized.Name, normalized.ID), nil)
		}
		if _, exists := seenNames[normalized.Name]; exists {
			return invalidArgument(fmt.Sprintf("subscription %q duplicates name", normalized.Name), nil)
		}
		seenIDs[normalized.ID] = struct{}{}
		seenNames[normalized.Name] = struct{}{}
		provider, subscription, err := mihomoNativeSubscriptionProvider(normalized)
		if err != nil {
			return err
		}
		p.Config.Subscriptions = append(p.Config.Subscriptions, subscription)
		p.Config.ProxyProviders[subscription.Name] = provider
	}
	return nil
}
