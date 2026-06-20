package mihomonative

import (
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

// DecodeMihomoNativeSettings decodes raw JSON into the mihomo native config and
// normalizes it, returning normalized defaults when raw is empty.
func DecodeMihomoNativeSettings(raw string) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	settings := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode mihomo native settings: %w", err)
		}
	}
	return NormalizeSettings(settings), nil
}
