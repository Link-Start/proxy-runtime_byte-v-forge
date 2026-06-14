package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a runtimeSettingsApplication) GetProxyRuntimeSettings(ctx context.Context) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	return a.usecase.Get(ctx)
}
