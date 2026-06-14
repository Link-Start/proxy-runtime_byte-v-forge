package mihomo

import "strings"

func newBaseRenderedConfig(opts renderOptions, gateway mihomoListener) mihomoConfig {
	return mihomoConfig{
		MixedPort:          gateway.Port,
		BindAddress:        gateway.Listen,
		AllowLAN:           true,
		Mode:               "rule",
		LogLevel:           "warning",
		ExternalController: strings.TrimSpace(opts.APIAddr),
		Secret:             strings.TrimSpace(opts.ControllerSecret),
		ExternalUI:         strings.TrimSpace(opts.DashboardDir),
		ExternalUIURL:      strings.TrimSpace(opts.DashboardURL),
		Authentication:     renderAuthentication(gateway.Users),
	}
}
