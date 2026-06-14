package mihomo

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func profileLayerGroup(name string, layer sourceplane.EgressProfileLayer, groupType string) mihomoGroup {
	return mihomoGroup{
		Name:           name,
		Type:           profileGroupStrategy(groupType),
		URL:            firstNonEmpty(layer.HealthCheckURL, "https://www.gstatic.com/generate_204"),
		Interval:       seconds(layer.HealthInterval, 300),
		Timeout:        milliseconds(layer.HealthTimeout, 5000),
		Lazy:           true,
		ExpectedStatus: defaultExpectedStatus(layer.ExpectedStatus),
		Hidden:         true,
	}
}

func profileGroupStrategy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "url-test", "select":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "select"
	}
}
