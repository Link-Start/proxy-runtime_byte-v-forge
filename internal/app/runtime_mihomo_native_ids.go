package app

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func nativeStableID(prefix string, source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(source))
	return runtimeSafeID(prefix + "-" + hex.EncodeToString(sum[:])[:12])
}
