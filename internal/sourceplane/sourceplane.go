package sourceplane

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type Driver interface {
	Name() string
	Reconcile(ctx context.Context, cfg Config) ([]provider.Node, error)
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
	Providers           []SubscriptionProvider
	FixedProxies        []FixedProxy
	EgressProfiles      []EgressProfile
	Endpoint            Endpoint
	GroupStrategy       string
	HealthCheckURL      string
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type Endpoint struct {
	Addr     string
	Protocol string
}

type FixedProxy struct {
	ID          string
	DisplayName string
	URI         string
	RegionCodes []string
}

type SubscriptionProvider struct {
	ID             string
	DisplayName    string
	URL            string
	Path           string
	Filter         string
	ExcludeFilter  string
	Interval       time.Duration
	HealthCheckURL string
	HealthInterval time.Duration
	HealthTimeout  time.Duration
	HealthLazy     bool
	ExpectedStatus uint32
	RegionCodes    []string
	Headers        map[string][]string
}

type EgressProfile struct {
	ID          string
	DisplayName string
	Enabled     bool
	Line        EgressProfileLine
	Exit        EgressProfileExit
}

type EgressProfileLayer struct {
	Kind           string
	ProviderID     string
	SourceID       string
	NodeID         string
	HealthCheckURL string
	HealthInterval time.Duration
	HealthTimeout  time.Duration
	ExpectedStatus uint32
}

type EgressProfileLine = EgressProfileLayer

type EgressProfileExit = EgressProfileLayer

type Empty struct{}

func (Empty) Name() string                                               { return "none" }
func (Empty) Reconcile(context.Context, Config) ([]provider.Node, error) { return nil, nil }
func (Empty) SourceNodes(context.Context, string) ([]*proxyruntimev1.ProxySourceNode, error) {
	return nil, nil
}
func (Empty) ResolveNodePublicIP(context.Context, string, string, string) (string, error) {
	return "", nil
}
func (Empty) Stop()          {}
func (Empty) Status() Status { return Status{LastError: "disabled"} }
