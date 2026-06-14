package app

import "strings"

func dynamicLeaseNativeDialerProxy(profileID string, nativeConfig mihomoNativeConfigFile, resourceID string, nodeID string) string {
	fixedByID, fixedByName := currentFixedProxyIndexes(nativeConfig)
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
	subscriptionByID, subscriptionByName := currentSubscriptionIndexes(nativeConfig)
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
