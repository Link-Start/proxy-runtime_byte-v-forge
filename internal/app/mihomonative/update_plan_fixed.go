package mihomonative

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (p *UpdatePlan) addFixedProxies(current ConfigFile, fixedProxies []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) error {
	currentByID, currentByName := CurrentFixedProxyIndexes(current)
	seenIDs := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	for _, item := range fixedProxies {
		normalized := NormalizeFixedProxy(FixedProxyFromProto(item), currentByName)
		if existing := currentByID[normalized.ID]; existing.ID != "" && existing.Name != normalized.Name {
			p.ResourceReplacements[existing.Name] = ResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		}
		p.ResourceReplacements[normalized.Name] = ResourceReplacement{ResourceID: normalized.ID, FixedProxy: true}
		if _, exists := seenIDs[normalized.ID]; exists {
			return fmt.Errorf("fixed proxy %q duplicates id %q", normalized.Name, normalized.ID)
		}
		if _, exists := seenNames[normalized.Name]; exists {
			return fmt.Errorf("fixed proxy %q duplicates name", normalized.Name)
		}
		seenIDs[normalized.ID] = struct{}{}
		seenNames[normalized.Name] = struct{}{}
		proxy, err := proxyFromURI(normalized.Name, normalized.URI)
		if err != nil {
			return err
		}
		normalized.Type = jsonStringValue(proxy["type"])
		p.Config.FixedProxies = append(p.Config.FixedProxies, normalized)
		p.Config.Proxies = append(p.Config.Proxies, proxy)
	}
	return nil
}
