package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
