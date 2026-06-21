package accountproxy

import (
	"strconv"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/random"
)

func passthroughUsername(base string, _ *proxygatewayv1.ProxySessionPolicy, _ string) string {
	return strings.TrimSpace(base)
}

func dashUsername(base string, pairs ...string) string {
	parts := []string{strings.TrimSpace(base)}
	for index := 0; index+1 < len(pairs); index += 2 {
		key := strings.TrimSpace(pairs[index])
		value := strings.TrimSpace(pairs[index+1])
		if key != "" && value != "" {
			parts = append(parts, key, value)
		}
	}
	return strings.Join(parts, "-")
}

func stickyMinutesString(policy *proxygatewayv1.ProxySessionPolicy) string {
	return strconv.Itoa(policyStickyMinutes(policy))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func numericSessionID() (string, error) {
	return random.String("0123456789", 8)
}
