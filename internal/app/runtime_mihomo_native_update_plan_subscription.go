package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
