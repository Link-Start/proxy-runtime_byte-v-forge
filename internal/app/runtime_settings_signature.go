package app

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
)

func runtimeSettingsSignature(settings *runtimeSettingsFile, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry) string {
	data, _ := protojsonx.Marshal(normalizeRuntimeSettingsWithProviders(settings, ipFraudProviders, ipGeoProviders))
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
