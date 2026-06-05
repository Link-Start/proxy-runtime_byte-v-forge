package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

func routeRuntimeKind(name string) proxyruntimev1.ProxyRouteRuntimeKind {
	if strings.EqualFold(name, "mihomo") {
		return proxyruntimev1.ProxyRouteRuntimeKind_PROXY_ROUTE_RUNTIME_KIND_MIHOMO
	}
	return proxyruntimev1.ProxyRouteRuntimeKind_PROXY_ROUTE_RUNTIME_KIND_UNSPECIFIED
}

func statusString(status dataplane.Status) string {
	if !status.Running {
		return firstNonEmpty(status.LastError, "stopped")
	}
	return "running"
}
