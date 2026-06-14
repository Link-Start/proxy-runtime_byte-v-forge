package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/random"
)

func PathInPrefix(requestPath string, prefix string) bool {
	requestPath = strings.TrimRight(strings.TrimSpace(requestPath), "/")
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/")
}

func ForwardedProto(req *http.Request) string {
	if req == nil {
		return "http"
	}
	if value := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto")); value != "" {
		return value
	}
	if req.TLS != nil {
		return "https"
	}
	return "http"
}

func RequestID(req *http.Request) string {
	if req != nil {
		for _, header := range []string{"X-Request-Id", "X-Request-ID", "X-Correlation-Id"} {
			if value := strings.TrimSpace(req.Header.Get(header)); value != "" {
				return value
			}
		}
	}
	value, err := random.Hex(8)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return value
}
