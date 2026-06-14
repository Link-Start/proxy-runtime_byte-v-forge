package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type mihomoNativeSettingsRepository interface {
	loadMihomoNative(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	saveMihomoNative(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error
}
