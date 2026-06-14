package app

import (
	"context"
	"fmt"
)

func (r *Runtime) dynamicLeaseDialerProxy(ctx context.Context, settings *runtimeSettingsFile, profileID string) (string, map[string]string, error) {
	settings = normalizeRuntimeSettings(settings)
	profiles := dynamicLeaseLineProfiles(settings, profileID)
	if len(profiles) == 0 {
		return "", nil, nil
	}
	nativeSettings, err := r.settings.loadMihomoNative(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("load mihomo native settings for lease line: %w", err)
	}
	nativeConfig, err := mihomoNativeConfigFileFromSettings(nativeSettings)
	if err != nil {
		return "", nil, fmt.Errorf("render mihomo native settings for lease line: %w", err)
	}
	for _, profile := range profiles {
		dialer, labels, err := dynamicLeaseProfileDialerProxy(profile, nativeConfig)
		if err != nil {
			return "", nil, err
		}
		if dialer != "" {
			return dialer, labels, nil
		}
	}
	return "", nil, nil
}
