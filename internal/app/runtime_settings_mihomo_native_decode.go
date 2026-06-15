package app

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

func decodeMihomoNativeSettings(raw string) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	settings := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode mihomo native settings: %w", err)
		}
	}
	return mihomonative.NormalizeSettings(settings), nil
}
