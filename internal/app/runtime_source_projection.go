package app

import (
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type sourcePlaneConfigInput struct {
	Settings          *runtimeSettingsFile
	Pool              []provider.Node
	LocalAddr         string
	HealthCheckURL    string
	HealthCheckPeriod time.Duration
	HealthCheckWait   time.Duration
	DashboardDir      string
	DashboardURL      string
	ProxyUsers        []config.ProxyUserRoute
}

func sourcePlaneDataPlaneConfig(input sourcePlaneConfigInput) dataplane.Config {
	return dataplane.Config{
		EgressProfiles:    sourcePlaneEgressProfiles(input.Settings),
		Endpoint:          sourceplane.Endpoint{Addr: input.LocalAddr, Protocol: "socks5"},
		HealthCheckURL:    input.HealthCheckURL,
		HealthCheckPeriod: input.HealthCheckPeriod,
		HealthCheckWait:   input.HealthCheckWait,
		DashboardDir:      input.DashboardDir,
		DashboardURL:      input.DashboardURL,
		Pool:              input.Pool,
		ProxyUsers:        sourcePlaneProxyUserRoutesWithConfigured(input.Settings, input.ProxyUsers),
	}
}
