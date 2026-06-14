package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type mihomoNativeUpdateRepository interface {
	loadMihomoNative(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	saveMihomoNative(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error
	replaceMihomoResourceRefs(context.Context, map[string]mihomoNativeResourceReplacement) (bool, error)
}

type mihomoNativeUpdateDependencies struct {
	Repository mihomoNativeUpdateRepository
	ConfigDir  string
	AfterApply func()
}
