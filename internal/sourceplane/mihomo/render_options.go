package mihomo

import (
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type renderOptions struct {
	EgressProfiles      []sourceplane.EgressProfile
	Endpoint            sourceplane.Endpoint
	ConfigDir           string
	NativeConfig        mihomoNativeConfig
	APIAddr             string
	ControllerSecret    string
	DashboardDir        string
	DashboardURL        string
	HealthCheckURL      string
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	BasePool            []provider.Node
	AvailableProxies    map[string]struct{}
	AvailableProviders  map[string]struct{}
	ProxyUsers          []dataplane.ProxyUserRoute
	SessionRoutes       []dataplane.SessionRoute
	ProfileGroups       map[string]string
}

type renderOptionsInput struct {
	Config          sourceplane.Config
	Endpoint        sourceplane.Endpoint
	ConfigDir       string
	NativeConfig    mihomoNativeConfig
	BaseConfig      dataplane.Config
	SessionRoutes   []dataplane.SessionRoute
	DashboardDir    string
	DashboardURL    string
	ControllerAddr  string
	ControllerToken string
}

func newRenderOptions(input renderOptionsInput) renderOptions {
	return renderOptions{
		EgressProfiles:      input.Config.EgressProfiles,
		Endpoint:            input.Endpoint,
		ConfigDir:           input.ConfigDir,
		NativeConfig:        input.NativeConfig,
		APIAddr:             input.ControllerAddr,
		ControllerSecret:    input.ControllerToken,
		DashboardDir:        input.DashboardDir,
		DashboardURL:        input.DashboardURL,
		HealthCheckURL:      input.Config.HealthCheckURL,
		HealthCheckInterval: input.Config.HealthCheckInterval,
		HealthCheckTimeout:  input.Config.HealthCheckTimeout,
		BasePool:            input.BaseConfig.Pool,
		ProxyUsers:          input.BaseConfig.ProxyUsers,
		SessionRoutes:       input.SessionRoutes,
	}
}
