package lease

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type DynamicIPSelection struct {
	Plan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	Endpoint accountproxy.Gateway
}
