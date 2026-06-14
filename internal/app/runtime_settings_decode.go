package app

import (
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

func decodeRuntimeSettings(raw string) (*runtimeSettingsFile, error) {
	settings := &runtimeSettingsFile{}
	if raw != "" {
		if err := protojsoncodec.Unmarshal([]byte(raw), settings); err != nil {
			return nil, fmt.Errorf("decode runtime settings: %w", err)
		}
	}
	return normalizeRuntimeSettings(settings), nil
}
