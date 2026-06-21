package mihomonative

import (
	"fmt"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"
)

// DecodeMihomoNativeSettings decodes raw JSON into the mihomo native config and
// normalizes it, returning normalized defaults when raw is empty.
func DecodeMihomoNativeSettings(raw string) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	settings := &proxygatewayv1.ProxyGatewayMihomoNativeConfig{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode mihomo native settings: %w", err)
		}
	}
	return NormalizeSettings(settings), nil
}
