package mihomonative

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func StableID(prefix string, source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(source))
	return appcore.RuntimeSafeID(prefix + "-" + hex.EncodeToString(sum[:])[:12])
}
