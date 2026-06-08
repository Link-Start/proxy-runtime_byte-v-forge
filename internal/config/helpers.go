package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func normalizeConfigToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func envString(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}

func envStringDefault(name string, fallback string) string {
	if value := envString(name); value != "" {
		return value
	}
	return fallback
}

func envDurationSeconds(name string, fallback time.Duration) (time.Duration, error) {
	value := envString(name)
	if value == "" {
		return fallback, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be integer seconds: %w", name, err)
	}
	return time.Duration(seconds) * time.Second, nil
}

func envList(name string) []string {
	value := envString(name)
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func proxyExitGeoURLs(name string) []string {
	values := envList(name)
	if len(values) == 0 {
		values = []string{"https://ipv4.icanhazip.com", "https://4.ident.me", "https://ifconfig.me/ip", "https://api.ipify.org?format=json", "https://checkip.global.api.aws/"}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}
