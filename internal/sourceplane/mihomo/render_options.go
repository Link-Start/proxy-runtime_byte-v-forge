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
