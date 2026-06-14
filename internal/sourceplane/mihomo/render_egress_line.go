package mihomo

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func renderEgressProfileLine(opts renderOptions, profileID string, line sourceplane.EgressProfileLine) (renderedProfileLayer, error) {
	switch strings.TrimSpace(line.Kind) {
	case "", "direct":
		return renderedProfileLayer{}, nil
	case "mihomo_node":
		return renderMihomoNativeProfileTarget(opts, profileID, line)
	default:
		return renderedProfileLayer{}, fmt.Errorf("unsupported line kind %q", line.Kind)
	}
}
