package domain

import (
	"crypto/sha256"
	"encoding/hex"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"
	"github.com/byte-v-forge/proxy-gateway/internal/ipgeo"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

// RuntimeSettingsSignature returns a stable content hash of the normalized
// runtime settings, used to detect provider-affecting changes.
func RuntimeSettingsSignature(settings *proxygatewayv1.ProxyGatewayPersistentSettings, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry) string {
	data, _ := protojsoncodec.Marshal(kernel.NormalizeRuntimeSettingsWithProviders(settings, ipFraudProviders, ipGeoProviders))
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
