package store

import (
	"github.com/byte-v-forge/proxy-runtime/internal/random"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func GeneratedSecretID(provider string, purpose string) (string, error) {
	suffix, err := random.Hex(12)
	if err != nil {
		return "", err
	}
	return secretref.StableID("proxy-runtime-secret", provider, purpose, suffix), nil
}
