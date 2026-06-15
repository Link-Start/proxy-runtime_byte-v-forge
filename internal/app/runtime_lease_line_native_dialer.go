package app

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
)

func dynamicLeaseNativeDialerProxy(profileID string, nativeConfig mihomonative.ConfigFile, resourceID string, nodeID string) string {
	fixedByID, fixedByName := mihomonative.CurrentFixedProxyIndexes(nativeConfig)
	for _, key := range []string{resourceID, nodeID, mihomoNodeDialerProxyName(resourceID, nodeID)} {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if proxy := fixedByID[key]; proxy.Name != "" {
			return proxy.Name
		}
		if proxy := fixedByName[key]; proxy.Name != "" {
			return proxy.Name
		}
	}
	subscriptionByID, subscriptionByName := mihomonative.CurrentSubscriptionIndexes(nativeConfig)
	for _, key := range []string{resourceID, mihomoNodeResourcePrefix(nodeID)} {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if subscriptionByID[key].Name != "" || subscriptionByName[key].Name != "" {
			return dynamicLeaseProfileLineGroupName(profileID)
		}
	}
	return ""
}
