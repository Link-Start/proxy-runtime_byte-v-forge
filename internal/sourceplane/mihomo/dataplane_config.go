package mihomo

import (
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func sourceConfigFromDataPlane(cfg dataplane.Config) sourceplane.Config {
	return sourceplane.Config{
		EgressProfiles:      cfg.EgressProfiles,
		Endpoint:            cfg.Endpoint,
		HealthCheckURL:      cfg.HealthCheckURL,
		HealthCheckInterval: cfg.HealthCheckPeriod,
		HealthCheckTimeout:  cfg.HealthCheckWait,
	}
}

func cloneDataPlaneConfig(cfg dataplane.Config) dataplane.Config {
	cfg.EgressProfiles = cloneEgressProfiles(cfg.EgressProfiles)
	cfg.Pool = cloneProviderNodes(cfg.Pool)
	cfg.ProxyUsers = append([]dataplane.ProxyUserRoute(nil), cfg.ProxyUsers...)
	return cfg
}
