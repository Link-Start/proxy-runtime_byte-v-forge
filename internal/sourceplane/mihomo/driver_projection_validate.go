package mihomo

import (
	"errors"
	"fmt"
)

func validateRenderedMihomoConfig(config mihomoConfig) error {
	if config.MixedPort <= 0 || config.MixedPort > 65535 {
		return fmt.Errorf("invalid rendered mihomo mixed port %d", config.MixedPort)
	}
	if config.Mode == "" {
		return errors.New("rendered mihomo mode is required")
	}
	if config.LogLevel == "" {
		return errors.New("rendered mihomo log level is required")
	}
	if len(config.Rules) == 0 {
		return errors.New("rendered mihomo rules are required")
	}
	return nil
}
