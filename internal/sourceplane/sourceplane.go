package sourceplane

import (
	"context"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type Driver interface {
	Name() string
	Reconcile(ctx context.Context, cfg Config) ([]provider.Node, error)
	Stop()
	Status() Status
}

type Status struct {
	Running    bool
	ConfigPath string
	LastError  string
}

type Config struct {
	EgressProfiles      []EgressProfile
	Endpoint            Endpoint
	HealthCheckURL      string
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
}

type Endpoint struct {
	Addr     string
	Protocol string
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
	ResourceID     string
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
func (Empty) Stop()                                                      {}
func (Empty) Status() Status                                             { return Status{LastError: "disabled"} }
