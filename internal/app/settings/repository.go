package settings

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type Repository interface {
	View(context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error)
}
