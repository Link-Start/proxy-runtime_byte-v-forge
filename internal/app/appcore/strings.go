package appcore

import (
	"fmt"
	"strings"
)

// FirstNonEmpty returns the first trimmed non-empty value, or "".
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

// CleanList trims, drops empties and de-duplicates values, preserving order.
func CleanList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// RuntimeSafeID keeps only [A-Za-z0-9_-], replacing other runes with '-' and
// trimming leading/trailing '-'.
func RuntimeSafeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var out strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			out.WriteRune(r)
			continue
		}
		out.WriteByte('-')
	}
	return strings.Trim(out.String(), "-")
}

// ShortHash returns an 8-hex-digit FNV-style hash of value.
func ShortHash(value string) string {
	return fmt.Sprintf("%08x", HashModulo(value, 0xffffffff))
}

// HashModulo returns an FNV-1a hash of value, reduced modulo when modulo > 0.
func HashModulo(value string, modulo uint32) uint32 {
	var h uint32 = 2166136261
	for _, ch := range []byte(value) {
		h ^= uint32(ch)
		h *= 16777619
	}
	if modulo > 0 {
		return h % modulo
	}
	return h
}
