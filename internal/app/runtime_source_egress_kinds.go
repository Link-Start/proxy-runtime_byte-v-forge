package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func egressProfileLineKind(kind proxyruntimev1.EgressProfileLineKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE:
		return "mihomo_node"
	default:
		return "direct"
	}
}

func egressProfileExitKind(kind proxyruntimev1.EgressProfileExitKind) string {
	switch kind {
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DIRECT:
		return "direct"
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
		return "static_ip"
	case proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
		return "dynamic_ip"
	default:
		return ""
	}
}
