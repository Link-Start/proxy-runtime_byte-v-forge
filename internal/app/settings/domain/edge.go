package domain

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func EdgeCanaryEnabled(settings *proxygatewayv1.ProxyEdgeCanarySettings) bool {
	return settings != nil && settings.GetEnabled()
}
