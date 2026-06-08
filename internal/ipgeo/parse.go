package ipgeo

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func parseGeo(payload map[string]any) *proxyruntimev1.ProxyExitGeo {
	return &proxyruntimev1.ProxyExitGeo{
		CountryCode: stringValue(payload, "country_code", "country", "countryCode", "location.country_code"),
		Region:      stringValue(payload, "region", "region_code", "region_name", "regionName", "state", "location.state"),
		City:        stringValue(payload, "city", "city_name", "location.city"),
	}
}

func stringValue(payload map[string]any, paths ...string) string {
	for _, path := range paths {
		value, ok := pathValue(payload, path)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if text := strings.TrimSpace(typed); text != "" {
				return text
			}
		case float64, bool:
			return strings.TrimSpace(fmt.Sprint(typed))
		}
	}
	return ""
}

func pathValue(payload map[string]any, path string) (any, bool) {
	var current any = payload
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
