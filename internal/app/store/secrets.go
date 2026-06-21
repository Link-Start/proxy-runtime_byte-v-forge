package store

import (
	"github.com/byte-v-forge/proxy-gateway/internal/random"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"
)

func GeneratedSecretID(provider string, purpose string) (string, error) {
	suffix, err := random.Hex(12)
	if err != nil {
		return "", err
	}
	return secretref.StableID("proxy-gateway-secret", provider, purpose, suffix), nil
}
