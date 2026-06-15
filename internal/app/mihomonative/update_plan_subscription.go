package mihomonative

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (p *UpdatePlan) addSubscriptions(current ConfigFile, subscriptions []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) error {
	currentByID, currentByName := CurrentSubscriptionIndexes(current)
	seenIDs := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	for _, item := range subscriptions {
		normalized := NormalizeSubscription(SubscriptionFromProto(item), currentByName)
		if existing := currentByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			p.ResourceReplacements[existing.Name] = ResourceReplacement{ResourceID: normalized.ID}
		}
		p.ResourceReplacements[normalized.Name] = ResourceReplacement{ResourceID: normalized.ID}
		if _, exists := seenIDs[normalized.ID]; exists {
			return fmt.Errorf("subscription %q duplicates id %q", normalized.Name, normalized.ID)
		}
		if _, exists := seenNames[normalized.Name]; exists {
			return fmt.Errorf("subscription %q duplicates name", normalized.Name)
		}
		seenIDs[normalized.ID] = struct{}{}
		seenNames[normalized.Name] = struct{}{}
		provider, subscription, err := SubscriptionProvider(normalized)
		if err != nil {
			return err
		}
		p.Config.Subscriptions = append(p.Config.Subscriptions, subscription)
		p.Config.ProxyProviders[subscription.Name] = provider
	}
	return nil
}
