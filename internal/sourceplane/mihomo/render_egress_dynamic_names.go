package mihomo

import (
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func profileDynamicProxyName(profileID string, index int, node provider.Node) string {
	return safeID(fmt.Sprintf("%s-dynamic-%d-%s", profileInternalGroupName(profileID), index, providerNodeName("provider-pool", node, index)))
}
