package store

import "strings"

// Runtime settings rows are keyed by these stable identifiers.
const (
	RuntimeSettingsKey      = "runtime"
	MihomoNativeSettingsKey = "mihomo_native"
)

// NormalizeID trims surrounding whitespace from a record identifier.
func NormalizeID(value string) string { return strings.TrimSpace(value) }
