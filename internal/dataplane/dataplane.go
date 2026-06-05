package dataplane

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type Driver interface {
	Name() string
	ReconcileBase(ctx context.Context, cfg Config) ([]provider.Node, error)
	UpsertSessionRoute(ctx context.Context, route SessionRoute) error
	DeleteSessionRoute(ctx context.Context, route SessionRoute) error
	SourceNodes(ctx context.Context, sourceID string) ([]*proxyruntimev1.ProxySourceNode, error)
	ResolveNodePublicIP(ctx context.Context, sourceID string, nodeID string, nodeDisplayName string) (string, error)
	Stop()
	Status() Status
}

type Status struct {
	Running    bool
	ConfigPath string
	LastError  string
}

type Config struct {
	SourceProviders   []sourceplane.SubscriptionProvider
	FixedProxies      []sourceplane.FixedProxy
	EgressProfiles    []sourceplane.EgressProfile
	Endpoint          sourceplane.Endpoint
	GroupStrategy     string
	HealthCheckURL    string
	HealthCheckPeriod time.Duration
	HealthCheckWait   time.Duration
	DashboardDir      string
	DashboardURL      string
	Common            *LocalService
	Local             LocalService
	Pool              []provider.Node
	DynamicViaCommon  bool
	ProxyUsers        []ProxyUserRoute
}

type LocalService struct {
	Name     string
	Addr     string
	Protocol string
	Username string
	Password string
	Route    string
}

type SessionRoute struct {
	SessionID string
	Listener  LocalService
	Pool      []provider.Node
}

type ProxyUserRoute struct {
	ID        string
	Username  string
	Password  string
	Route     string
	SourceID  string
	NodeID    string
	ProfileID string
}
