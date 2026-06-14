package dataplane

import (
	"context"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type Driver interface {
	Name() string
	ReconcileBase(ctx context.Context, cfg Config) ([]provider.Node, error)
	UpsertSessionRoute(ctx context.Context, route SessionRoute) error
	DeleteSessionRoute(ctx context.Context, route SessionRoute) error
	Stop()
	Status() Status
}

type Status struct {
	Running           bool
	ConfigPath        string
	DesiredConfigHash string
	AppliedConfigHash string
	LastError         string
}

type Config struct {
	EgressProfiles    []sourceplane.EgressProfile
	Endpoint          sourceplane.Endpoint
	HealthCheckURL    string
	HealthCheckPeriod time.Duration
	HealthCheckWait   time.Duration
	DashboardDir      string
	DashboardURL      string
	Pool              []provider.Node
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
	SessionID   string
	Listener    LocalService
	Pool        []provider.Node
	DialerProxy string
}

type ProxyUserRoute struct {
	ID        string
	Username  string
	Password  string
	Route     string
	ProfileID string
}
