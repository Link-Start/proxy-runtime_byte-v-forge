package lease

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"
)

type DynamicIPSelection struct {
	Plan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	Endpoint accountproxy.Gateway
}
