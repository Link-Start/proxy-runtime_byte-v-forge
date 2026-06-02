package app

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
)

func runtimeSettingsSignature(settings *runtimeSettingsFile, ipFraudProviders *ipfraud.Registry) string {
	data, _ := protojsonx.Marshal(normalizeRuntimeSettingsWithProviders(settings, ipFraudProviders))
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
