package app

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func updateMihomoNativeSettings(ctx context.Context, runtime *Runtime, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if runtime == nil {
		return nil, internalError("runtime is required", nil)
	}
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	currentView, err := runtime.settings.loadMihomoNative(ctx)
	if err != nil {
		return nil, internalError("load mihomo native settings", err)
	}
	current, err := mihomoNativeConfigFileFromSettings(currentView)
	if err != nil {
		return nil, internalError("load mihomo native settings", err)
	}
	next := mihomoNativeConfigFile{
		FixedProxies:   make([]mihomoNativeFixedProxy, 0, len(view.FixedProxies)),
		Proxies:        make([]map[string]any, 0, len(view.FixedProxies)),
		ProxyProviders: map[string]mihomoNativeProvider{},
		ProxyGroups:    preserveNativeGroups(current.ProxyGroups),
		Rules:          append([]string(nil), current.Rules...),
	}
	currentFixedByID, currentFixedByName := currentFixedProxyIndexes(current)
	resourceReplacements := map[string]mihomoNativeResourceReplacement{}
	seenFixedIDs := map[string]struct{}{}
	seenFixedNames := map[string]struct{}{}
	for _, item := range view.GetFixedProxies() {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(item), currentFixedByName)
		if existing := currentFixedByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			resourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		}
		resourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		if _, exists := seenFixedIDs[normalized.ID]; exists {
			return nil, invalidArgument(fmt.Sprintf("fixed proxy %q duplicates id %q", normalized.Name, normalized.ID), nil)
		}
		if _, exists := seenFixedNames[normalized.Name]; exists {
			return nil, invalidArgument(fmt.Sprintf("fixed proxy %q duplicates name", normalized.Name), nil)
		}
		seenFixedIDs[normalized.ID] = struct{}{}
		seenFixedNames[normalized.Name] = struct{}{}
		proxy, err := mihomoNativeProxyFromURI(normalized.Name, normalized.URI)
		if err != nil {
			return nil, invalidArgument(err.Error(), nil)
		}
		normalized.Type = jsonStringValue(proxy["type"])
		next.FixedProxies = append(next.FixedProxies, normalized)
		next.Proxies = append(next.Proxies, proxy)
	}
	currentSubscriptionsByID, currentSubscriptionsByName := currentSubscriptionIndexes(current)
	seenSubscriptionIDs := map[string]struct{}{}
	seenSubscriptionNames := map[string]struct{}{}
	for _, item := range view.GetSubscriptions() {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(item), currentSubscriptionsByName)
		if existing := currentSubscriptionsByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			resourceReplacements[existing.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		}
		resourceReplacements[normalized.Name] = mihomoNativeResourceReplacement{ResourceID: normalized.ID}
		if _, exists := seenSubscriptionIDs[normalized.ID]; exists {
			return nil, invalidArgument(fmt.Sprintf("subscription %q duplicates id %q", normalized.Name, normalized.ID), nil)
		}
		if _, exists := seenSubscriptionNames[normalized.Name]; exists {
			return nil, invalidArgument(fmt.Sprintf("subscription %q duplicates name", normalized.Name), nil)
		}
		seenSubscriptionIDs[normalized.ID] = struct{}{}
		seenSubscriptionNames[normalized.Name] = struct{}{}
		provider, subscription, err := mihomoNativeSubscriptionProvider(normalized)
		if err != nil {
			return nil, err
		}
		next.Subscriptions = append(next.Subscriptions, subscription)
		next.ProxyProviders[subscription.Name] = provider
	}
	if err := runtime.settings.saveMihomoNative(ctx, mihomoNativeSettingsFromConfig(next)); err != nil {
		return nil, internalError("save mihomo native settings", err)
	}
	if err := saveMihomoNativeConfig(runtime, next); err != nil {
		return nil, internalError("save mihomo native config", err)
	}
	_, err = runtime.settings.replaceMihomoResourceRefs(ctx, resourceReplacements)
	if err != nil {
		return nil, internalError("update mihomo native resource references", err)
	}
	runtime.exitCheckCache.clear()
	runtime.requestReconcile()
	return mihomoNativeSettings(ctx, runtime)
}
