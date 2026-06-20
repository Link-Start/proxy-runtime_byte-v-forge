package domain

import (
	"crypto/sha256"
	"encoding/hex"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

// RuntimeSettingsSignature returns a stable content hash of the normalized
// runtime settings, used to detect provider-affecting changes.
func RuntimeSettingsSignature(settings *proxyruntimev1.ProxyRuntimePersistentSettings, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry) string {
	data, _ := protojsoncodec.Marshal(kernel.NormalizeRuntimeSettingsWithProviders(settings, ipFraudProviders, ipGeoProviders))
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
